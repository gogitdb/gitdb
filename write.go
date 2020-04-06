package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func (g *gitdb) Insert(mo Model) error {

	if err := mo.BeforeInsert(); err != nil {
		return fmt.Errorf("Model.BeforeInsert failed: %s", err)
	}

	if err := mo.Validate(); err != nil {
		return fmt.Errorf("Model is not valid: %s", err)
	}

	m := wrap(mo)

	if err := m.GetSchema().Validate(); err != nil {
		return err
	}

	if err := g.flushQueue(); err != nil {
		log(err.Error())
	}
	return g.write(m)
}

func (g *gitdb) InsertMany(models []Model) error {
	//todo polish this up later
	if len(models) > 100 {
		return errors.New("max number of models InsertMany supports is 100")
	}

	tx := g.StartTransaction("InsertMany")
	var model Model
	for _, model = range models {
		//create a new variable to pass to function to avoid
		//passing pointer which will end up inserting the same
		//model multiple times
		m := model
		f := func() error { return g.Insert(m) }
		tx.AddOperation(f)
	}
	return tx.Commit()
}

func (g *gitdb) queue(m Model) error {

	if g.writeQueue == nil {
		g.writeQueue = map[string]Model{}
	}

	g.writeQueue[ID(m)] = m
	return nil
}

func (g *gitdb) flushQueue() error {

	for id, model := range g.writeQueue {
		log("Flushing: " + id)

		err := g.write(model)
		if err != nil {
			logError(err.Error())
			return err
		}

		delete(g.writeQueue, id)
	}

	return nil
}

func (g *gitdb) write(m Model) error {

	if _, err := os.Stat(g.fullPath(m)); err != nil {
		err := os.MkdirAll(g.fullPath(m), 0755)
		if err != nil {
			return fmt.Errorf("failed to make dir %s: %w", g.fullPath(m), err)
		}
	}

	schema := m.GetSchema()
	blockFilePath := g.blockFilePath(schema.name(), schema.block)
	dataBlock, err := g.loadBlock(blockFilePath, schema.name())
	if err != nil {
		return err
	}

	logTest(fmt.Sprintf("Size of block before write - %d", dataBlock.size()))

	//...append new record to block
	newRecordBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	mID := ID(m)

	//construct a commit message
	commitMsg := "Inserting " + mID + " into " + schema.blockID()
	if _, err := dataBlock.get(mID); err == nil {
		commitMsg = "Updating " + mID + " in " + schema.blockID()
	}

	newRecordStr := string(newRecordBytes)
	//encrypt data if need be
	if m.ShouldEncrypt() {
		newRecordStr = encrypt(g.config.EncryptionKey, newRecordStr)
	}

	dataBlock.add(m.GetSchema().recordID(), newRecordStr)

	g.events <- newWriteBeforeEvent("...", mID)
	if err := g.writeBlock(blockFilePath, dataBlock); err != nil {
		return err
	}

	log(fmt.Sprintf("autoCommit: %v", g.autoCommit))

	g.commit.Add(1)
	g.events <- newWriteEvent(commitMsg, blockFilePath, g.autoCommit)
	logTest("sent write event to loop")
	g.updateIndexes(schema.name(), newRecord(mID, newRecordStr))

	//block here until write has been committed
	g.waitForCommit()

	return nil
}

func (g *gitdb) waitForCommit() {
	if g.autoCommit {
		logTest("waiting for gitdb to commit changes")
		g.commit.Wait()
	}
}

func (g *gitdb) writeBlock(blockFile string, block *block) error {
	g.writeMu.Lock()
	defer g.writeMu.Unlock()

	blockBytes, fmtErr := json.MarshalIndent(block, "", "\t")
	if fmtErr != nil {
		return fmtErr
	}

	return ioutil.WriteFile(blockFile, blockBytes, 0744)
}

func (g *gitdb) Delete(id string) error {
	return g.dodelete(id, false)
}

func (g *gitdb) DeleteOrFail(id string) error {
	return g.dodelete(id, true)
}

func (g *gitdb) dodelete(id string, failNotFound bool) error {

	dataset, block, _, err := ParseID(id)
	if err != nil {
		return err
	}

	blockFilePath := g.blockFilePath(dataset, block)
	err = g.delByID(id, dataset, blockFilePath, failNotFound)

	if err == nil {
		logTest("sending delete event to loop")
		g.commit.Add(1)
		g.events <- newDeleteEvent("Deleting "+id+" in "+blockFilePath, blockFilePath, g.autoCommit)
		g.waitForCommit()
	}

	return err
}

func (g *gitdb) delByID(id string, dataset string, blockFile string, failIfNotFound bool) error {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	dataBlock := newBlock(dataset)
	err := g.readBlock(blockFile, dataBlock)
	if err != nil {
		return err
	}

	if err := dataBlock.delete(id); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	//write undeleted records back to block file
	return g.writeBlock(blockFile, dataBlock)
}
