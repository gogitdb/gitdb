GitDB
=====

[![Go Report Card](https://goreportcard.com/badge/github.com/fobilow/gitdb?style=flat-square)](https://goreportcard.com/report/github.com/fobilow/gitdb)
[![Coverage](https://codecov.io/gh/fobilow/gitdb/branch/develop/graph/badge.svg)](https://codecov.io/gh/fobilow/gitdb)
[![Build Status Travis](https://img.shields.io/travis/fobilow/gitdb.svg?style=flat-square&&branch=master)](https://travis-ci.com/fobilow/gitdb)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/fobilow/gitdb)
[![Releases](https://img.shields.io/github/release/fobilow/gitdb/all.svg?style=flat-square)](https://github.com/fobilow/gitdb/releases)
[![LICENSE](https://img.shields.io/github/license/fobilow/gitdb.svg?style=flat-square)](https://github.com/fobilow/gitdb/blob/master/LICENSE)

## What is GitDB?

> GitDB is not a binary. It’s a library!

GitDB is a decentralized document database written in Go that uses [Git](https://git-scm.com/) under the hood to provide database-like functionalities via strictly defined interfaces. 

GitDB allows developers to create Models of objects in their application which implement a Model Interface that can access it's persistence features. This allows GitDB to work with these objects in database operations. 

## Why GitDB - motivation behind project?

- A need for a database that was quick and simple to set up
- A need for a database that was decentralized and each participating client in a system can store their data independent of other clients.


## Features

- Decentralized
- Document store
- Embedded into your go application
- Encryption (encrypt on write, decrypt on read)
- Record locking.
- Simple Indexing System
- Transactions
- Web UI 


## Project versioning

GitDB uses [semantic versioning](http://semver.org).
API should not change between patch and minor releases.
New minor versions may add additional features to the API.

## Table of Contents

  - [Getting Started](#getting-started)
    - [Installing](#installing)
    - [Configuration](#configuration)
    - [Opening a database](#opening-a-database)
    - [Inserting/Updating a record](#insertingupdating-a-record)
    - [Fetching a single record](#fetching-a-single-record)
    - [Fetching all records in a dataset](#fetching-all-records-in-a-dataset)
    - [Deleting a record](#deleting-a-record)
    - [Search for records](#search-for-records)
    - [Transactions](#transactions)
    - [Encryption](#encryption)
  - [Resources](#resources)
  - [Caveats & Limitations](#caveats--limitations)
  - [Reading the Source](#reading-the-source)
  <!-- - [Other Projects Using GitDB](#other-projects-using-gitdb) -->

## Getting Started

### Installing

To start using GitDB, install Go and run `go get`:

```sh
$ go get github.com/fobilow/gitdb/v2
```


### Configuration

Below are configuration options provided by GitDB
<table>
  <tr style="font-weight:bold;">
    <td>Name</td>
    <td width="600">Description</td>
    <td>Type</td>
    <td>Required</td>
    <td>Default</td>
  <tr>
  <tr>
    <td>DbPath</td>
    <td>Path on your machine where you want GitDB to create/clone your database</td>
    <td>string</td>
    <td>Y</td>
    <td>N/A</td>
  </tr>
  <tr>
    <td>ConnectionName</td>
    <td>Unique name for gitdb connection. Use this when opening multiple GitDB connections</td>
    <td>string</td>
    <td>N</td>
    <td>"default"</td>
  </tr>
  <tr>
    <td>OnlineRemote</td>
    <td>URL for remote git server you want GitDB to sync with e.g git@github.com:user/db.git or https://github.com/user/db.git.
    <p><strong>Note: The first time GitDB runs, it will automatically generate ssh keys and will automatically attempt to use this key to sync with the OnlineRemote,
    therefore ensure that the generated keys are added to this git server. The ssh keys can be found at <i>Config.DbPath/.gitdb/ssh</i></strong></p>
    </td>
    <td>string</td>
    <td>N</td>
    <td>""</td>
  </tr>
  <tr>
    <td>SyncInterval</td>
    <td>This controls how often you want GitDB to sync with the online remote</td>
    <td>time.Duration.</td>
    <td>N</td>
    <td>5s</td>
  </tr>
  <tr>
    <td>EncryptionKey</td>
    <td>16,24 or 32 byte string used to provide AES encryption for Models that implement ShouldEncrypt</td>
    <td>string</td>
    <td>N</td>
    <td>""</td>
  </tr>
  <tr>
    <td>User</td>
    <td>This specifies the user connected to the Gitdb and will be used to commit all changes to the database</td>
    <td>gitdb.User</td>
    <td>N</td>
    <td>ghost &#x3C;ghost@gitdb.local&#x3E;</td>
  </tr>
  <tr>
    <td>EnableUI</td>
    <td>Use this option to enable GitDB web user interface</td>
    <td>bool</td>
    <td>N</td>
    <td>false</td>
  </tr>
  <tr>
    <td>UIPort</td>
    <td>Use this option to change the default port which GitDB uses to serve it's web user interface</td>
    <td>int</td>
    <td>N</td>
    <td>4120</td>
  </tr>
  <tr>
    <td>Factory</td>
    <td>For backward compatibity with v1. In v1 GitDB needed a factory method to be able construct concrete Model for certain database operations.
    This has now been dropped in v2
    </td>
    <td>func(dataset string) gitdb.Model</td>
    <td>N</td>
    <td>nil</td>
  </tr>
  <tr>
    <td>Mock</td>
    <td>Flag used for testing apps. If true, will return a mock GitDB connection</td>
    <td>bool</td>
    <td>N</td>
    <td>false</td>
  </tr>
</table>

You can configure GitDB either using the constructor or constructing it yourself

```go	
cfg := gitdb.NewConfig(path)
//or
cfg := gitdb.Config{
  DbPath: path
}
```


<!-- This will retrieve the library and install the `gitdb` command line utility into
your `$GOBIN` path. -->

### Importing GitDB

To use GitDB as an embedded document store, import as:

```go
import "github.com/fobilow/gitdb/v2"

cfg := gitdb.NewConfig(path)
db, err := gitdb.Open(cfg)
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
  "github.com/fobilow/gitdb/v2"
)

func main() {
  
  cfg := gitdb.NewConfig("/tmp/data")
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

A Model is a struct that represents a record in GitDB. GitDB only works with models that implement the gidb.Model interface

gitdb.TimeStampedModel is a simple struct that allows you to easily add CreatedAt and UpdatedAt to all the Models in your application and will automatically time stamp them before persisting to GitDB. You can write your own base Models to embed common fields across your application Models

```go
type BankAccount struct {
  //TimeStampedModel will add CreatedAt and UpdatedAt fields this Model
  gitdb.TimeStampedModel 
  AccountType         string
  AccountNo           string
  Currency            string
  Name                string
}

func (b *BankAccount) GetSchema() *gitdb.Schema {
  //Dataset Name
  name := "Accounts"
  //Block ID
  block := b.CreatedAt.Format("200601")
  //Record ID
  record := b.AccountNo

  //Indexes speed up searching
  indexes := make(map[string]interface{})
  indexes["AccountType"] = b.AccountType
 
  return gitdb.NewSchema(name, block, record, indexes)
}

func (b *BankAccount) Validate() error            { return nil }
func (b *BankAccount) IsLockable() bool           { return false }
func (b *BankAccount) ShouldEncrypt() bool        { return false }
func (b *BankAccount) GetLockFileNames() []string { return []string{} }

...
  
```

### Inserting/Updating a record
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
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
    log.Println(err)
  }

  //get account id
  log.Println(gitdb.Id(account))

  //update account name
  account.Name = "Bar Foo"
  err = db.Insert(account)
  if err != nil {
    log.Println(err)
  }
}
```

### Fetching a single record
```go
package main
import (
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  //model to passed to Get to store result 
  var account BankAccount()
  err = db.Get("Accounts/202003/0123456789", &account)
  if err != nil {
    log.Println(err)
  }
}
```
### Fetching all records in a dataset
```go
package main

import (
  "fmt"
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
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
    log.Print(fmt.Sprintf("%s-%s", gitdb.ID(b), b.AccountNo))
  }
}

```
### Deleting a record
```go
package main

import (
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
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
  "fmt"
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  //Find all records that have savings account type
  searchParam := &db.SearchParam{Index: "AccountType", Value: "Savings"}
  records, err := dbconn.Search("Accounts", []*db.SearchParam{searchParam}, gitdb.SearchEquals)
  if err != nil {
    log.Println(err.Error())
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
  "github.com/fobilow/gitdb/v2"
)

func main() {
  cfg := gitdb.NewConfig("/tmp/data")
  db, err := gitdb.Open(cfg)
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  func accountUpgradeFuncOne() error { println("accountUpgradeFuncOne..."); return nil }
  func accountUpgradeFuncTwo() error { println("accountUpgradeFuncTwo..."); return errors.New("accountUpgradeFuncTwo failed") }
  func accountUpgradeFuncThree() error { println("accountUpgradeFuncThree"); return nil }

  tx := db.StartTransaction("AccountUpgrade")
  tx.AddOperation(accountUpgradeFuncOne)
  tx.AddOperation(accountUpgradeFuncTwo)
  tx.AddOperation(accountUpgradeFuncThree)
  terr := tx.Commit()
  if terr != nil {
    log.Print(terr)
  }
}
```

### Encryption

GitDB suppports AES encryption and is done on a Model level, which means you can have a database with different Models where some are encrypted and others are not. To encrypt your data, your Model must implement `ShouldEncrypt()` to return true and you must set `gitdb.Config.EncryptionKey`. For maximum security set this key to a 32 byte string to select AES-256 

```go
package main

import (
  "log"
  "github.com/fobilow/gitdb/v2"
)

func main(){
  cfg := gitdb.NewConfig("/tmp/data")
  cfg.EncryptionKey = "a_32_bytes_string_for_AES-256"
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

  //Insert will encrypt the account
  err = db.Insert(account)
  if err != nil {
    log.Println(err)
  }

  //Get will automatically decrypt account
  var account BankAccount()
  err = db.Get("Accounts/202003/0123456789", &account)
  if err != nil {
    log.Println(err)
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

Below is a list of public, open source projects that use GitDB:

* VogueHotel - Uses Gitdb as the default database backend.


If you are using GitDB in a project please send a pull request to add it to the list. -->