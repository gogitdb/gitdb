package db

type Operation func() error

type Transaction struct {
	name string
	operations []Operation
}

func (t *Transaction) Commit() error {
	autoCommit = false
	for _, o := range t.operations {
		if err := o(); err != nil {
			log("Reverting transaction: "+err.Error())
			gitUndo()
			autoCommit = true
			return err
		}
	}
	commitMsg := "Committing transaction: " + t.name
	events <- newWriteEvent(commitMsg, ".")
	autoCommit = true
	return nil
}

func (t *Transaction) AddOperation(o Operation) {
	t.operations = append(t.operations, o)
}

func NewTransaction(name string) *Transaction {
	return &Transaction{name: name}
}