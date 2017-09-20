# Google Datastore ORM

## Quick Start
### Configure datastore
```go
if err := db.Config(projectID); err != nil {
  panic("Datastore error: ", err)
}
```

### Define table struct for database (one struct is one table)
#### Merchant Table
```go
// Merchant : Merchant is root
type Merchant struct {
	ID          int64             `datastore:"-"` // load table id
	Key         *datastore.Key    `datastore:"__key__"` // load table key
	Name        string              
	Type        string            
}

// LoadKey : populate property after Load
func (x *Merchant) LoadKey(k *datastore.Key) error {
	x.Key = k
	x.ID = k.ID
	return nil
}

// Load : load property
func (x *Merchant) Load(ps []datastore.Property) error {
	return datastore.LoadStruct(x, ps)
}

// Save : manipulate property
func (x *Merchant) Save() ([]datastore.Property, error) {
	return datastore.SaveStruct(x)
}
```
#### User Table
```go
// User : User kind parent is Merchant
type User struct {
	Key        *datastore.Key    `datastore:"__key__"` // load table key
	Name       string              
	Age        int64            
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

  if err := db.Kind("Merchant").Create(merchant, nil); err != nil {
    log.Println(err) // fail to create record
  }
```
### Read Record
* __Get Parent Key__
```go
  // Example
  user := new(user)
  if err := db.Kind("User").Find(key, user); err != nil {
    log.Println(err) // error while retrieving record
  }

  parentKey := user.Key.Parent // get the merchant key because user parent is merchant
```
* __Get Single Record__
```go
  // Example 1
  merchant := new(Merchant)
  if err := db.Kind("Merchant").First(merchant); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 2
  merchant := new(Merchant)
  if err := db.Kind("Merchant").Where("Name =", "Hello World").First(merchant); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 3
  user := new(user)
  if err := db.Kind("User").Where("Name =", "myz").Where("Age =", 22).Ancestor(parentKey).First(user); err != nil {
    log.Println(err) // error while retrieving record
  }
```

* __Get Multiple Record__
```go
  // Example 1
  merchants := new([]Merchant)
  if err := db.Kind("Merchant").Limit(10).Get(merchants); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 2
  merchants := new([]Merchant)
  if err := db.Kind("Merchant").Where("Name =", "Hello World").Get(merchants); err != nil {
    log.Println(err) // error while retrieving record
  }

  // Example 3
  users := new([]user)
  if err := db.Kind("User").Where("Name =", "myz").Where("Age =", 22).Ancestor(parentKey).Get(users); err != nil {
    log.Println(err) // error while retrieving record
  }
```

* __Find Record By Key__
```go
  // Example
  user := new(User)
  if err := db.Kind("User").Find(key, user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }
```

* __Pagination Record__
```go
  p := db.Pagination{
    Limit:  2,
		Cursor: "", // pass the cursor that generate by the query so that it will display the next record
  }
	
  // Example
  users := new([]user)
  if err := db.Kind("User").Where("Name =", "myz").Ancestor(parentKey).Paginate(&p).Get(users); err != nil {
    log.Println(err) // error while retrieving record
  }

	cursor := p.Cursor // pass this cursor to db.Pagination struct so that it will display the next two record
```

* __Count Record__
```go
  // Example
  user := new([]User)
  n, err := db.Kind("User").Count(key, user); 
	if err != nil {
    log.Println(err) // error while count record
  }
	log.Println(n) // number of record
```

### Update Record
```go
  // Example
  user := new(user)
  if err := db.Kind("User").Where("Name =", "myz").Where("Age =", 22).First(user); err != nil {
    log.Println(err) // error while retrieving record
  }
  if user.Key != nil {
    user.Name = "Hello World"
    user.Age = 20
    if err := db.Kind("User").Update(user); err != nil {
    log.Println(err) // fail to update record
   }
  }
```

### Delete Record
```go
  // Example
  user := new(user)
  if err := db.Kind("User").Find(key, user); err != nil {
    log.Println(err) // error while retrieving record or record not found
  }

  if err := db.Kind("User").Delete(user.Key); err != nil {
    log.Println(err) // fail to delete record
  }
```