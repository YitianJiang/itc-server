# itc-server

## 运行程序

```bash
cd ~/go/src/code.byted.org/clientQA/itc-server
bash run.sh
```

## 加入我们

### 拉取代码

```Bash
mkdir -p ~/go/src/code.byted.org/clientQA
cd ~/go/src/code.byted.org/clientQA
git clone git@code.byted.org:clientQA/itc-server.git
```

## 一些建议

使用数据库时，请采用`database.DB()`的方式，该方法会返回一个全局的数据库句柄，全局数据库句柄会在`main`函数中初始化。采用这种方式有两个好处：

- 减少代码冗余，数据库句柄仅初始化一次，全局可用。

- 数据库句柄限制为包级可见，可避免不知情的修改，提高代码的健壮性。

如果想在数据库中`插入/删除`记录，请考虑使用下面这两个函数，目前可以在`database`包中找到它们：

```Go
// DeleteDBRecord deletes all eligbile record in the given database table.
// Note that the parameter table should be the instance or pointer of the struct
// mapping to the table and table name will be invalid.
// If the model has a DeletedAt field, it will get a soft delete ability
// automatically, which means the record will not be permanently removed
// from the database, rather the DeletedAt' value will be set to the current time.
func DeleteDBRecord(db *gorm.DB, table interface{}, sieve map[string]interface{}) error {

	if err := db.Debug().Delete(table, sieve).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}

// InsertDBRecord inserts record into database if it is valid.
// Note that the parameter record must be the pointer of the struct mapping to
// the target table.
// If the operation succeeded, record.ID is exact the id of the inserted record.
func InsertDBRecord(db *gorm.DB, record interface{}) error {

	if err := db.Debug().Create(record).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}
```