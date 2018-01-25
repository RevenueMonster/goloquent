# Google Datastore ORM 
## Installation
* __Clone the project to src folder__
```bash
  // dependency
  $ go get -u github.com/go-sql-driver/mysql
  $ go get -u cloud.google.com/go/datastore

  $ go get -u github.com/RevenueMonster/goloquent
```
* __Import the library__
```go
  import "github.com/revenuemonster/goloquent"
```

## Quick Start
### Connect to database
```go
// Connect to google datastore 
db, err := goloquent.Open("datastore", "projectid")
if err != nil {
  panic("Connection error: ", err)
}

// Connect to mysql 
db, err := goloquent.Open("mysql", "username@/projectid")
if err != nil {
  panic("Connection error: ", err)
}
```
#### User Table
```go
// User : User kind parent is Merchant
type User struct {
	Key        *datastore.Key `datastore:"__key__"` // load table key
	Name       string         `datastore:"name"`
	Age        int64          `datastore:"age"`
}

// LoadKey : populate property after Load
func (x *User) LoadKey(k *datastore.Key) error {
	x.Key = k
	return nil
}

// Load : load property
func (x *User) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(x, ps)
}

// Save : manipulate property
func (x *User) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(x)
}
```
### Create Record
```go
  // Example
  merchant := new(Merchant)
  merchant.Name = "Hello World"
  merchant.Type = "Type 1"

  // OR 
  var merchant *Merchant
  merchant.Name = "Hello World"
  merchant.Type = "Type 1"

  // Create without parent key
  if err := db.Table("Merchant").Create(merchant, nil); err != nil {
    log.Println(err) // fail to create record
  }

  // Create with parent key
  parentKey := datastore.NameKey("Parent", "value", nil)
  if err := db.Table("Merchant").Create(merchant, parentKey); err != nil {
    log.Println(err) // fail to create record
  }

  // Create with self generate key
  key := datastore.NameKey("Merchant", "uniqueID", nil)
  if err := db.Table("Merchant").Create(merchant, key); err != nil {
    log.Println(err) // fail to create record
  }
```
### Retrieve Record
* __Get Single Record using Primary Key__
```go
  // Example
  primaryKey := datastore.IDKey("User", 1039213, nil)
  user := new(User)
  if err := db.Table("User").Find(primaryKey, user); err != nil {
    log.Println(err) // error while retrieving record
  }
```
* __Get Single Record__
```go
  // Example 1
  user := new(User)
  if err := db.Table("User").First(user); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 2
  user := new(User)
  if err := db.Table("User").Where("Email", "=", "admin@hotmail.com").First(user); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 3
  parentKey := datastore.IDKey("Parent", 1093, nil)
  user := new(User)
  if err := db.Table("User").Where("Age", "=", 22).Ancestor(parentKey).Order("-CreatedDateTime").First(user); err != nil {
    log.Println(err) // error while retrieving record
  }
```

* __Get Multiple Record__
```go
  // Example 1
  merchants := new([]Merchant)
  if err := db.Table("Merchant").Limit(10).Get(merchants); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 2
  merchants := new([]*Merchant)
  if err := db.Table("Merchant").Where("Name", "=", "Hello World").Get(merchants); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 3
  users := new([]user)
  if err := db.Table("User").Where("Name", "=", "myz").Where("Age", "=", 22).Ancestor(parentKey).Get(users); err != nil {
    log.Println(err) // error while retrieving record
  }
```
* __Get Record with Ordering__
```go
  // Ascending order
  users := new([]*User)
  if err := db.Table("User").Order("CreatedDateTime").Get(users); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Descending order
  users := new([]*User)
  if err := db.Table("User").Order("-CreatedDateTime").Get(users); err != nil {
    log.Println(err) // error while retrieving record
  }
```

* __Pagination Record__
```go
  p := goloquent.Pagination{
    Limit:  10,
    Cursor: "", // pass the cursor that generate by the query so that it will display the next record
  }
	
  // Example
  users := new([]*User)
  if err := db.Table("User").Ancestor(parentKey).Order("-CreatedDateTime").Paginate(&p, users); err != nil {
    log.Println(err) // error while retrieving record
  }
```

* __Count Record__
```go
  // Example
  n, err := db.Table("User").Count()
  if err != nil {
    log.Println(err) // error while count record
  }
```

### Delete Record
```go
  // Example
  user := new(user)
  if err := db.Table("User").First(user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }

  if err := db.Table("User").Delete(user.Key); err != nil {
    log.Println(err) // fail to delete record
  }
```
###Transaction
```go
  // Example
  if err := db.RunInTransaction(func(txn *goloquent.Connection) error {
    user := new(User)
    if err := txn.Table("User").Create(user, nil); err != nil {
      return err
    }
    return nil
  }); err != nil {
    log.Println(err)
  }
```
###MySQL Exclusive
* __Database Migration__
```go
  // Example
  user := new(User)
  if err := db.Table("User").Migrate(user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }
```

* __Drop Database__
```go
  // This will throw error if table is not exist
  if err := db.Table("User").Drop(); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }

  // Drop table if exists
  if err := db.Table("User").DropIfExists(); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }
```

* __Change Schema__


```go
type User struct {
  Name  string `goloquent:"longtext"` // By default string is varchar(255)
  Int   int    `goloquent:"unsigned"` // Unsigned option only applicable for int data type
  Email string `goloquent:"unique"`   // Make column Email as unique field
}
```

| Data Type               | Schema        |
| :---------------------- | :------------ |
| *datastore.Key          | varchar(767)  |
| string                  | varchar(255)  |
| []byte                  | blob          |
| bool                    | boolean       |
| float32                 | decimal(10,2) |
| float64                 | decimal(10,2) |
| int, int8, int16, int32 | int           |
| int64                   | big integer   |
| slice or array          | text          |
| struct                  | text          |
| time.Time               | datetime      |


```go
By default, $Key, $Parent, $PrimaryKey are reserved words
```

* __Filter Query__
```go
  // Update single record
  user := new(User)
  if err := db.Table("User").Where("Status", "IN", []string{"active", "pending"}).First(user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }

  // Get record with like
  if err := db.Table("User").Where("Name", "LIKE", "%name%").First(user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }
```

* __Update Query__
```go
  // Update single record
  user := new(User)
  user.Key = datastore.IDKey("User", 167393, nil)
  user.Name = "Test"
  if err := db.Table("User").Update(user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }

  // Update multiple record
  if err := db.Table("User").Where("Age", ">", 10).Update(map[string]interface{}{
    "Name": "New Name",
    "Email": "email@gmail.com",
  }); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }
```
