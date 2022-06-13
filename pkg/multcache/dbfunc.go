package multcache

import (
	"database/sql"
	"gorm.io/gorm"
	"log"
)

// SqlQueryCount sql query count opention Template
func SqlQueryCount(args ...interface{}) (interface{}, error) {
	db, ok := args[0].(*gorm.DB)
	if !ok {
		log.Fatalf("error type!")
	}
	table, ok := args[1].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	sqlCmd, ok := args[2].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	condition, ok := args[3].(uint32)
	if !ok {
		log.Fatalf("error type!")
	}
	var count int64
	dbTx := db.Table(table).Where(sqlCmd, condition).Count(&count)
	// TODO：ensure error type while count==0
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	if dbTx.Error == errNotFoundInSql {
		return 0, nil
	}
	if dbTx.Error != nil {
		return 0, dbTx.Error
	}
	return count, nil
}

// SqlBatchQuery sql batch query count opention Template
func SqlBatchQuery(args ...interface{}) (interface{}, error) {
	db, ok := args[0].(*gorm.DB)
	if !ok {
		log.Fatalf("error type!")
	}
	table, ok := args[1].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	limit, ok := args[2].(int)
	if !ok {
		log.Fatalf("error type!")
	}
	offset, ok := args[3].(int)
	if !ok {
		log.Fatalf("error type!")
	}
	orderCondition, ok := args[4].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	var accounts interface{}
	dbTx := db.Table(table).Limit(int(limit)).Offset(int(offset)).Order(orderCondition).Find(&accounts)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == errRowsAffectedNull {
		return nil, nil
	}
	return accounts, nil
}

func SqlBatchQueryWithWhere(args ...interface{}) (interface{}, error) {
	db, ok := args[0].(*gorm.DB)
	if !ok {
		log.Fatalf("error type!")
	}
	table, ok := args[1].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	where, ok := args[2].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	whereCondition, ok := args[3].(sql.NamedArg)
	if !ok {
		log.Fatalf("error type!")
	}
	limit, ok := args[4].(int)
	if !ok {
		log.Fatalf("error type!")
	}
	offset, ok := args[5].(int)
	if !ok {
		log.Fatalf("error type!")
	}
	orderCondition, ok := args[6].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	var resultList interface{}
	dbTx := db.Table(table).Where(where, whereCondition).Limit(int(limit)).Offset(int(offset)).Order(orderCondition).Find(&resultList)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == errRowsAffectedNull {
		return nil, nil
	}
	return resultList, nil
}

func SqlQueryCountNamed(args ...interface{}) (interface{}, error) {
	db, ok := args[0].(*gorm.DB)
	if !ok {
		log.Fatalf("error type!")
	}
	table, ok := args[1].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	where, ok := args[2].(string)
	if !ok {
		log.Fatalf("error type!")
	}
	whereCondition, ok := args[3].(sql.NamedArg)
	if !ok {
		log.Fatalf("error type!")
	}
	var count int64
	dbTx := db.Table(table).Where(where, whereCondition).Count(&count)
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	if dbTx.Error == errNotFoundInSql {
		return 0, nil
	}
	if dbTx.Error != nil {
		return 0, dbTx.Error
	}
	return count, nil
}
