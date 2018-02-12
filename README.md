# Google Datastore ORM 
Inspired by Laravel Eloquent [https://laravel.com/docs/5.5/eloquent](https://laravel.com/docs/5.5/eloquent)

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
    Key             *datastore.Key `goloquent:"__key__"` // load table key
    Name            string
    CountryCode     string
    PhoneNumber     string
    Age             int64          `goloquent:",unsigned"`
    CreatedDateTime time.Time
    UpdatedDateTime time.Time
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

### Helper 
```go
    goloquent.SetDebug(true) // Enable debug mode in goloquent

    // StringPrimaryKey is to get key value in string form
    key := datastore.NameKey("User", "mjfFgYnxBS", nil)
    fmt.Println(goloquent.StringPrimaryKey(key)) // "mjfFgYnxBS"

    key = datastore.IDKey("User", int64(2305297334603281546), nil)
    fmt.Println(goloquent.StringPrimaryKey(key)) // "2305297334603281546"
```

### Create Record
```go
    // Example
    user := new(User)
    user.Name = "Hello World"
    user.Age = 18

    // OR 
    var user *User
    user.Name = "Hello World"
    user.Age = 18

    // Create without parent key
    if err := db.Table("User").Create(user, nil); err != nil {
        log.Println(err) // fail to create record
    }

    // Create with parent key
    parentKey := datastore.NameKey("Parent", "value", nil)
    if err := db.Table("User").Create(merchant, parentKey); err != nil {
        log.Println(err) // fail to create record
    }

    // Create with self generate key
    key := datastore.NameKey("User", "uniqueID", nil)
    if err := db.Table("User").Create(user, key); err != nil {
        log.Println(err) // fail to create record
    }
```

### Upsert Record
```go
    // Example
    user := new(User)
    user.Name = "Hello World"
    user.Age = 18

    // Update if key exists, else create the user record
    parentKey := datastore.NameKey("Parent", "value", nil)
    if err := db.Table("User").Upsert(merchant, parentKey); err != nil {
        log.Println(err) // fail to create record
    }

    // Upsert with self generate key
    key := datastore.NameKey("User", "uniqueID", nil)
    if err := db.Table("User").Upsert(user, key); err != nil {
        log.Println(err) // fail to create record
    }
```

### Retrieve Record
* __Get Single Record using Primary Key__
```go
    // Example
    primaryKey := datastore.IDKey("User", int64(2305297334603281546), nil)
    user := new(User)
    if err := db.Table("User").Find(primaryKey, user); err != nil {
        log.Println(err) // error while retrieving record
    }

    if err := db.Table("User").
        Where("Status", "=", "ACTIVE").
        Find(primaryKey, user); err != goloquent.ErrNoSuchEntity {
        // if record not found using primary key, error ErrNoSuchEntity will throw instead
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

    if user.Key != nil { // if have record
        fmt.Println("Have record")
    } else {
        fmt.Println("Doesnt't have record")
    }

    // Example 2
    user := new(User)
    if err := db.Table("User").
        Where("Email", "=", "admin@hotmail.com").
        First(user); err != nil {
        log.Println(err) // error while retrieving record
    }

    // Example 3
    parentKey := datastore.IDKey("Parent", 1093, nil)
    user := new(User)
    if err := db.Table("User").
        Ancestor(parentKey).
        Where("Age", "=", 22).
        Order("-CreatedDateTime").
        First(user); err != nil {
        log.Println(err) // error while retrieving record
    }
```

* __Get Multiple Record__
```go
    // Example 1
    users := new([]User)
    if err := db.Table("User").
        Limit(10).
        Get(users); err != nil {
        log.Println(err) // error while retrieving record
    }

    // Example 2
    users := new([]*User)
    if err := db.Table("User").
        Where("Name", "=", "Hello World").
        Get(users); err != nil {
        log.Println(err) // error while retrieving record
    }

    // Example 3
    users := new([]user)
    if err := db.Table("User").
        Ancestor(parentKey).
        Where("Name", "=", "myz").
        Where("Age", "=", 22).
        Get(users); err != nil {
        log.Println(err) // error while retrieving record
    }
```
* __Get Record with Ordering__
```go
    // Ascending order
    users := new([]*User)
    if err := db.Table("User").
        Order("CreatedDateTime").
        Get(users); err != nil {
        log.Println(err) // error while retrieving record
    }

    // Descending order
    if err := db.Table("User").
        Order("-CreatedDateTime").
        Get(users); err != nil {
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
    if err := db.Table("User").
        Ancestor(parentKey).
        Order("-CreatedDateTime").
        Paginate(&p, users); err != nil {
        log.Println(err) // error while retrieving record
    }

    // ***************** OR ********************
    p := &goloquent.Pagination{
        Limit:  10, // number of records in each page
        Cursor: "EhQKCE1lcmNoYW50EK3bueKni5eNIxIWCgxMb3lhbHR5UG9pbnQaBkZrUUc4eA", // pass the cursor to get next record set
    }

    users := new([]*User)
    if err := db.Table("User").
        Ancestor(parentKey).
        Order("-CreatedDateTime").
        Paginate(p, users); err != nil {
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

    n, err := db.Table("User").Where("Age", ">", 10).Count()
    if err != nil {
        log.Println(err)
    }
```

### Delete Record
* __Delete using Primary Key__
```go
    // Example
    if err := db.Table("User").Delete(user.Key); err != nil {
        log.Println(err) // fail to delete record
    }
```
* __Delete using Where statement__
```go
    // Delete user table record which account type not equal to "PREMIUM" or "MONTLY"
    if err := db.Table("User").
        Where("AccountType", "!=", []string{
        "PREMIUM", "MONTLY",
        }).
        Delete(); err != nil {
        log.Println(err) // fail to delete record
    }
```

* __Soft Delete__
```go
    type User struct {
        Key  *datastore.Key `datastore:"__key__"` // primary key
        Name string 
        goloquent.SoftDelete // User struct will using SoftDelete when deleting  
    }

    user := new(User)
        if err := db.Table("User").First(user); err != nil {
        log.Println(err) // error while retrieving record or record not found
    }

    if err := db.Table("User").SoftDelete(user.Key); err != nil {
        log.Println(err) // fail to delete record
    }
```

### Transaction 
```go
    // Example
    if err := db.RunInTransaction(func(txn *goloquent.Connection) error {
        user := new(User)

        if err := txn.Table("User").Create(user, nil); err != nil {
            return err // return any err to rollback the transaction
        }

        return nil // return nil to commit the transaction
    }); err != nil {
        log.Println(err)
    }
```

### MySQL Exclusive
* __Database Migration__
```go
    // Example
    user := new(User)
    if err := db.Table("User").Migrate(user); err != nil {
        log.Println(err)
    }
```
* __Unique Index__
```go
    // Create unique Index
    if err := db.Table("User").UniqueIndex("CountryCode", "PhoneNumber"); err != nil {
        log.Println(err)
    }

    // Drop Unique Index
    if err := db.Table("User").DropUniqueIndex("CountryCode", "PhoneNumber"); err != nil {
        log.Println(err)
    }

    // Drop PrimaryKey
    if err := db.Table("User").DropUniqueIndex("__key__"); err != nil {
        log.Println(err)
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

* __Table Locking (only effective inside RunInTransaction)__
```go
    // Example
    userKey := datastore.IDKey("User", int64(4645436182170916864), nil)
    if err := db.RunInTransaction(func(txn *goloquent.Connection) error {
        user := new(User)

        if err := txn.Table("User").
            LockForUpdate(). // Lock record for update
            Find(userKey, user); err != nil {
            return err
        }

        if err := txn.Table("User").Update(user); err != nil {
            return err // return any err to rollback the transaction
        }

        return nil // return nil to commit the transaction
    }); err != nil {
        log.Println(err)
    }
```

* __Sum Column__
```go
    // Example
    n, err := db.Table("User").Sum("Age")
    if err != nil {
        log.Println(err) // error while sum record
    }
```

* __Filter Query__
```go
    // Update single record
    user := new(User)
    if err := db.Table("User").
        Where("Status", "IN", []string{"active", "pending"}).
        First(user); err != nil {
        log.Println(err) // error while retrieving record or record not found
    }

    // Get record with like
    if err := db.Table("User").
        Where("Name", "LIKE", "%name%").
        First(user); err != nil {
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
    if err := db.Table("User").
        Where("Age", ">", 10).
        Update(map[string]interface{}{
            "Name": "New Name",
            "Email": "email@gmail.com",
            "UpdatedDateTime": time.Now().UTC(),
        }); err != nil {
        log.Println(err) // error while retrieving record or record not found
    }
```

* __Extra Schema Option__
```go
type User struct {
    Name  string `goloquent:",longtext"` // Using `TEXT` datatype instead of `VARCHAR(255)` by default
    Age   int    `goloquent:",unsigned"` // Unsigned option only applicable for int data type
    Email string `goloquent:",unique"`   // Make column `Email` as unique field
    DefaultAddress struct {
        AddressLine1 string // `DefaultAddress.AddressLine1`
        AddressLine2 string // `DefaultAddress.AddressLine2`
        PostCode     int    // `DefaultAddress.PostCode`
        City         string // `DefaultAddress.City`
        State        string // `DefaultAddress.State`
        Country      string 
    } `goloquent:",flatten"` // Flatten the struct field
}
```

| Data Type               | Schema                    | Default Value       |
| :---------------------- | :------------------------ | :-------------------|
| *datastore.Key          | varchar(20), varchar(767) |                     |
| datastore.GeoPoint      | varchar(50)               | {Lat: 0, Lng: 0}    |
| string                  | varchar(255)              | ""                  |
| []byte                  | mediumblob                |                     |
| bool                    | boolean                   | false               |
| float32                 | decimal(10,2)             | 0                   |
| float64                 | decimal(10,2)             | 0                   |
| int, int8, int16, int32 | int                       | 0                   |
| int64                   | big integer               | 0                   |
| slice or array          | text                      | ""                  |                 
| struct                  | text                      | ""                  |
| time.Time               | datetime                  | 0001-01-01 00:00:00 |


**$Key**, **$Parent**, **$PrimaryKey** and **DeletedAt** are reserved words, please avoid to use these words as your column name


