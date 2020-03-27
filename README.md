GitDB
=====

[![Go Report Card](https://goreportcard.com/badge/github.com/fobilow/gitdb?style=flat-square)](https://goreportcard.com/report/github.com/etcd-io/bbolt)
[![Coverage](https://codecov.io/gh/fobilow/gitdb/branch/develop/graph/badge.svg)](https://codecov.io/gh/fobilow/gitdb)
[![Build Status Travis](https://img.shields.io/travis/fobilow/gitdb.svg?style=flat-square&&branch=master)](https://travis-ci.com/fobilow/gitdb)
<!-- [![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/etcd-io/bbolt) -->
<!-- [![Releases](https://img.shields.io/github/release/etcd-io/bbolt/all.svg?style=flat-square)](https://github.com/etcd-io/bbolt/releases) -->
<!-- [![LICENSE](https://img.shields.io/github/license/etcd-io/bbolt.svg?style=flat-square)](https://github.com/etcd-io/bbolt/blob/master/LICENSE) -->

## What is GitDB?

GitDB is to Go what Mnesia is to Erlang

- Decentralized database
- Document database
- Embedded into the go application
- Supports Encryption (encrypt on write, decrypt on read)
- Supports multiple data formats e.g Json, Bson (more to come...)
- Supports record locking.
- Simple Indexing System
- Decentralization is achieved with git
- Data is stored in blocks (blocks can be dynamic or static)
- A block can contain many records...


## Project versioning

GitDB uses [semantic versioning](http://semver.org).
API should not change between patch and minor releases.
New minor versions may add additional features to the API.

## Table of Contents

  - [Getting Started](#getting-started)
    - [Installing](#installing)
    - [Opening a database](#opening-a-database)
    - [Inserting/Updating a record](#insertingupdating-a-record)
    - [Fetching a single record](#fetching-a-single-record)
    - [Fetching all records in a dataset](#fetching-all-records-in-a-dataset)
    - [Deleting a record](#deleting-a-record)
    - [Search for records](#search-for-records)
    - [Transactions](#transactions)
  - [Resources](#resources)
  - [Caveats & Limitations](#caveats--limitations)
  - [Reading the Source](#reading-the-source)
  <!-- - [Other Projects Using GitDB](#other-projects-using-gitdb) -->

## Getting Started

### Installing

To start using GitDB, install Go and run `go get`:

```sh
$ go get github.com/fobilow/gitdb
```

<!-- This will retrieve the library and install the `gitdb` command line utility into
your `$GOBIN` path. -->

### Importing GitDB

To use GitDB as an embedded document store, import as:

```go
import "github.com/fobilow/gitdb"

config := gitdb.NewConfig(path)
db, err := gitdb.Open(path)
if err != nil {
  log.Fatal(err)
}
defer db.Close()
```


### Opening a database
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main() {
  
  config := gitdb.NewConfig("/tmp/data")
  // Open will create or clone down a git repo 
  // in configured path if it does not exist.
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  ...
}
```

### Models

A model represents a record in GitDB. GitDB only works with models that implement the gidb.Model interface

BaseModel is a sub type provided by gitdb to standardize and simplify the creation of Models in your application. It provides 4 standard fields to every Model in your application and error reporting feature of gitdb. It provides a partial implementation of the Model Interface which can be completed by composition with an application Model.

```go
type BankAccount struct {
  gitdb.BaseModel //using BaseModel by composition to simplify code
  AccountType         string
  AccountNo           string
  Currency            string
  Name                string
}

func (b *BankAccount) GetSchema() *gitdb.Schema {
  //Dataset Name
  name := func() string {return "Accounts"}
  //Block ID
  block := func() string {return b.CreatedAt.Format("200601")}
  //Record ID
  record := func() string {return b.AccountNo}

  //Indexes speed up searching
  indexes := func() map[string]interface{} {
     indexes := make(map[string]interface{})

     indexes["AccountType"] = b.AccountType
     return indexes
  }

  return gitdb.NewSchema(name, block, record, indexes)
}
  
```

### Inserting/Updating a record
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main(){
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  //populate model
  account := &BankAccount()
  account.AccountNo = "0123456789"
  account.AccountType = "Savings"
  account.Currency = "GBP"
  account.Name = "Foo Bar"

  err = db.Insert(account)
  if err != nil {
    fmt.Println(err)
  }

  //update account name
  account.Name = "Bar Foo"
  err = db.Insert(account)
  if err != nil {
    log.Error(err)
  }
}
```

### Fetching a single record
```go
package main
import (
  "log"
  "github.com/fobilow/gitdb"
)

func main(){
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  //model to passed to Get to store result 
  account := &BankAccount()
  err = db.Get("Accounts/202003/0123456789", account)
  if err != nil {
    log.Error(err)
  }
}
```
### Fetching all records in a dataset
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main(){
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  records, err := db.Fetch("Accounts")
  if err != nil {
    log.Print(err)
    return
  }

  accounts := []*BankAccount{}
  for _, r := range records {
    b := &BankAccount{}
    r.Hydrate(b)
    accounts = append(accounts, b)
    log.Print(fmt.Sprintf("%s-%s", b.ID, b.CreatedAt))
  }
}

```
### Deleting a record
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main(){
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  err := db.Delete("Accounts/202003/0123456789")
  if err != nil {
    log.Print(err)
  }
}
```

### Search for records
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main(){
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  searchParam := &db.SearchParam{Index: "AccountType", Value: "Savings"}
  records, err := dbconn.Search("Accounts", []*db.SearchParam{searchParam}, db.SEARCH_MODE_EQUALS)
  if err != nil {
    fmt.Println(err.Error())
    return
  } 

  accounts := []*BankAccount{}
  for _, r := range records {
    b := &BankAccount{}
    r.Hydrate(b)
    accounts = append(accounts, b)
    log.Print(fmt.Sprintf("%s-%s", b.ID, b.CreatedAt))
  }
}
```

### Transactions
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb"
)

func main() {
  config := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  func accountUpgradeFuncOne() error { println("accountUpgradeFuncOne..."); return nil }
  func accountUpgradeFuncTwo() error { println("accountUpgradeFuncTwo..."); return errors.New("accountUpgradeFuncTwo failed") }
  func accountUpgradeFuncThree() error { println("accountUpgradeFuncThree"); return nil }

  t := db.NewTransaction("AccountUpgrade")
  t.AddOperation(accountUpgradeFuncOne)
  t.AddOperation(accountUpgradeFuncTwo)
  t.AddOperation(accountUpgradeFuncThree)
  terr := t.Commit()
  if terr != nil {
    log.Print(terr)
  }
}
```

## Resources

For more information on getting started with Gitdb, check out the following articles:
* [Gtidb - an embedded distributed document database for Go](https://docs.google.com/document/d/1OPalq-7J_Vo_uks35up_4a_V0BsRrpwL1J4JG6ZFoEs/edit?usp=sharing) by Oke Ugwu


## Caveats & Limitations

It's important to pick the right tool for the job and GitDB is no exception.
Here are a few things to note when evaluating and using GitDB:

* GitDB is good for systems where data producers are indpendent. 
* GitDB currently depends on the git binary to work 

## Reading the Source

GitDB is a relatively small code base (<5KLOC) for an embedded, distributed,
document database so it can be a good starting point for people
interested in how databases work.

The best places to start are the main entry points into GitDB:

- `Open()` - Initializes the reference to the database. It's responsible for
  creating the database if it doesn't exist and pulling down existing database
  if an online remote is specified.

If you have additional notes that could be helpful for others, please submit
them via pull request.


<!-- ## Other Projects Using GitDB

Below is a list of public, open source projects that use Bolt:

* VogueHotel - Uses Gitdb as the default database backend.


If you are using Gitdb in a project please send a pull request to add it to the list. -->