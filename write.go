package gitdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bouggo/log"
	"github.com/gogitdb/gitdb/v2/internal/crypto"
	"github.com/gogitdb/gitdb/v2/internal/db"
)

func (g *gitdb) Insert(mo Model) error {
	m := wrap(mo)

	if err := m.Validate(); err != nil {
		return err
	}

	if err := m.BeforeInsert(); err != nil {
		return fmt.Errorf("Model.BeforeInsert failed: %s", err)
	}

	if err := m.GetSchema().Validate(); err != nil {
		return err
	}

	return g.insert(m)
}

func (g *gitdb) InsertMany(models []Model) error {
	tx := g.StartTransaction("InsertMany")
	for _, model := range models {
		//create a new variable to pass to function to avoid
		//passing pointer which will end up inserting the same
		//model multiple times
		m := model
		f := func() error { return g.Insert(m) }
		tx.AddOperation(f)
	}
	return tx.Commit()
}

func (g *gitdb) insert(m Model) error {
	if !g.isRegistered(m.GetSchema().dataset) {
		return ErrInvalidDataset
	}

	if _, err := os.Stat(g.fullPath(m)); err != nil {
		err := os.MkdirAll(g.fullPath(m), 0755)
		if err != nil {
			return fmt.Errorf("failed to make dir %s: %w", g.fullPath(m), err)
		}
	}

	schema := m.GetSchema()
	blockFilePath := g.blockFilePath(schema.name(), schema.block)
	dataBlock, err := g.loadBlock(blockFilePath)
	if err != nil {
		return err
	}

	log.Test(fmt.Sprintf("Size of block before write - %d", dataBlock.Len()))

	//...append new record to block
	newRecordBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	mID := ID(m)

	//construct a commit message
	commitMsg := "Inserting " + mID
	if _, err := dataBlock.Get(mID); err == nil {
		commitMsg = "Updating " + mID
	}

	newRecordStr := string(newRecordBytes)
	//encrypt data if need be
	if m.ShouldEncrypt() {
		newRecordStr = crypto.Encrypt(g.config.EncryptionKey, newRecordStr)
	}

	dataBlock.Add(mID, newRecordStr)

	g.events <- newWriteBeforeEvent("...", mID)
	if err := g.writeBlock(blockFilePath, dataBlock); err != nil {
		return err
	}

	log.Info(fmt.Sprintf("autoCommit: %v", g.autoCommit))

	g.commit.Add(1)
	g.events <- newWriteEvent(commitMsg, blockFilePath, g.autoCommit)
	log.Test("sent write event to loop")
	g.updateIndexes(dataBlock)

	//block here until write has been committed
	g.waitForCommit()

	return nil
}

func (g *gitdb) waitForCommit() {
	if g.autoCommit {
		log.Test("waiting for gitdb to commit changes")
		g.commit.Wait()
	}
}

func (g *gitdb) writeBlock(blockFile string, block *db.Block) error {
	g.writeMu.Lock()
	defer g.writeMu.Unlock()

	blockBytes, fmtErr := json.MarshalIndent(block, "", "\t")
	if fmtErr != nil {
		return fmtErr
	}

	//update cache
	if g.loadedBlocks != nil {
		g.loadedBlocks[blockFile] = block
	}
	return ioutil.WriteFile(blockFile, blockBytes, 0744)
}

func (g *gitdb) Delete(id string) error {
	return g.doDelete(id, false)
}

func (g *gitdb) DeleteOrFail(id string) error {
	return g.doDelete(id, true)
}

func (g *gitdb) doDelete(id string, failNotFound bool) error {

	dataset, block, _, err := ParseID(id)
	if err != nil {
		return err
	}

	blockFilePath := g.blockFilePath(dataset, block)
	err = g.delByID(id, blockFilePath, failNotFound)

	if err == nil {
		log.Test("sending delete event to loop")
		g.commit.Add(1)
		g.events <- newDeleteEvent(fmt.Sprintf("Deleting %s", id), blockFilePath, g.autoCommit)
		g.waitForCommit()
	}

	return err
}

func (g *gitdb) delByID(id string, blockFile string, failIfNotFound bool) error {

	if _, err := os.Stat(blockFile); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	dataBlock := db.LoadBlock(blockFile, g.config.EncryptionKey)
	if err := dataBlock.Delete(id); err != nil {
		if failIfNotFound {
			return errors.New("Could not delete [" + id + "]: record does not exist")
		}
		return nil
	}

	//write undeleted records back to block file
	return g.writeBlock(blockFile, dataBlock)
}
