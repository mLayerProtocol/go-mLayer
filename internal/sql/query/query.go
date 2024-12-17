package query

import (
	dbsql "database/sql"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/common/encoder"
	"github.com/mlayerprotocol/go-mlayer/common/utils"
	"github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/entities"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"gorm.io/gorm"
)

var logger = &log.Logger

func GetOne[T any, U any](filter T, data U) error {
	if err := utils.CheckEmpty(filter); err != nil {
		return err
	}
	// logger.Debugf("QUERY %v, value: %v, %v", filter, data)
	err := sql.SqlDb.Where(filter).Take(data).Error
	if err != nil {

		return err
	}
	return nil
}

func GetOneWithOr[T any, U any](where T, or T, data U) error {
	if err := utils.CheckEmpty(where); err != nil {
		return err
	}
	if err := utils.CheckEmpty(or); err != nil {
		return err
	}
	err := sql.SqlDb.Where(where).Or(or).Take(data).Error
	if err != nil {

		return err
	}
	return nil
}

func GetTx() *gorm.DB {
	return sql.SqlDb
}

func GetManyTx[T any](item T) *gorm.DB {
	return sql.SqlDb.Where(&item)
}

type Order string

const (
	OrderDec Order = "desc"
	OrderAsc Order = "asc"
)

func GetMany[T any, U any](item T, data *U, order *map[string]Order) error {
	tx := GetManyTx(item)
	if order != nil {
		for k := range *order {
			logger.Debugf("%s %s", k, (*order)[k])
			tx = tx.Order(fmt.Sprintf("%s %s", k, (*order)[k]))
		}
	}
	err := tx.Find(data).Error
	if err != nil {
		return err
	}
	return nil
}
func GetManyWithLimit[T any, U any](item T, data *U, order *map[string]Order, limit int, offset int) error {
	tx := GetManyTx(item)
	if order != nil {
		for k := range *order {
			logger.Debugf("%s %s", k, (*order)[k])
			tx = tx.Order(fmt.Sprintf("%s %s", k, (*order)[k]))
		}
	}
	if limit > 0 {
		tx.Limit(int(limit))
	}
	if offset > 0 {
		tx.Offset(offset)
	}
	err := tx.Find(data).Error
	if err != nil {
		return err
	}
	return nil
}


func GetWithIN[T any, U any, I any](item T, data *U, slice I) error {
	err := sql.SqlDb.Find(data, slice).Error

	if err != nil {
		return err
	}

	return nil
}

type Result struct {
	models.SubscriptionState
	Block uint
}

func GetSubscriptionsByBlock(subState entities.Subscription, from uint, to uint, block bool) ([]Result, error) {
	subStateTable := GetTableName(models.SubscriptionState{})
	subEventTable := GetTableName(models.SubscriptionEvent{})

	rows, err := sql.SqlDb.Model(&models.SubscriptionState{}).Select(fmt.Sprintf("%s.*, %s.block_number", subStateTable, subEventTable)).Where(models.SubscriptionState{Subscription: subState}).
		Joins(fmt.Sprintf("left join %s on  %s.event_hash = %s.hash", subEventTable, subStateTable, subEventTable)).
		Where(fmt.Sprintf("%s.block_number >= %d and %s.block_number < %d", subEventTable, from, subEventTable, to)).Rows()

	list := []Result{}
	for rows.Next() {
		rsl := Result{}
		rows.Scan(rsl)
		list = append(list, rsl)
	}
	if err != nil {
		return list, err
	}
	logger.Debug("RSL ", list)
	return list, nil
}

func GetManyGroupBy[T any, U any](item T, data *U, gb string) error {
	err := sql.SqlDb.Where(&item).Select(fmt.Sprintf("topic_events.*, count(%s) as total", gb)).Group(gb).Find(data).Error
	if err != nil {
		return err
	}

	return nil
}

// func GetDependentEvents [T any] (event entities.Event) (*[]T, error) {

// 	var data []T
// 	err := sql.SqlDb.Where(
// 		&models.AuthorizationEvent{Event: entities.Event{PreviousEvent: event.Hash}},
// 	).Or(&models.AuthorizationEvent{Event: entities.Event{AuthEvent: event.Hash}},
// 	// ).Or("? LIKE ANY (associations)", fmt.Sprintf("%%%s%%", event.Hash)
// 	).Find(&data).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &data, nil
// }

func GetOneState[T any](filter any, data *T) error {
		var model any
			switch val := filter.(type) {
				case entities.Subnet:
					// logger.Debug(val)
					model = models.SubnetState{Subnet: val}
				case entities.Authorization:
					model = models.AuthorizationState{Authorization: val}
				case entities.Topic:
					model = models.TopicState{Topic: val}
				case entities.Subscription:
					model = models.SubscriptionState{Subscription: val}
				case entities.Message:
					model = models.MessageState{Message: val}
			}
			// logger.Debugf("QUERY %v, value: %v, %v", filter, data)
	err := sql.SqlDb.Where(model).Take(data).Error
	if err != nil {

		return err
	}
	return nil
}



func SaveRecord[Model any](where Model,  createData *Model, updateData *Model, DB *gorm.DB) (model *Model, created bool, err error) {
	tx := DB
	if DB == nil {
		tx = sql.SqlDb
	}
	// dataByte, err := encoder.MsgPackStruct(event.Payload)
	// if err != nil {
	// 	return nil, false, err
	// }
	// authPayload := entities.ClientPayload{
	// 	Data: entities.Authorization{},
	// }
	// copier.Copy(&authPayload, &event.Payload)
	// data := models.AuthorizationEvent{
	// 	Event: *event,
	// }
	if err := utils.CheckEmpty(where); err != nil {
		return model, false, err
	}


	var result *gorm.DB
	// if utils.CheckEmpty(updateData) != nil {
		if err := utils.CheckEmpty(updateData); err != nil {
			result = tx.Where(where).FirstOrCreate(createData)
		} else {
			result = tx.Where(where).Assign(updateData).FirstOrCreate(createData)
			
		}
	// } else {
	// 	result = tx.Where(where).FirstOrCreate(createData)
	// }
	// if result.Error != nil {
	// 	if DB == nil {
	// 		tx.Rollback()
	// 	}
	// 	return model, false, result.Error
	// }
	// if DB == nil {
	// 	tx.Commit()
	// }
	return createData, result.RowsAffected > 0, nil
}

func SaveRecordWithMap[Model any](model *Model, where map[string]interface{}, data map[string]interface{}, update bool, DB *gorm.DB) (created bool, err error) {
	tx := DB
	if DB == nil {
		tx = sql.SqlDb
	}
	var result *gorm.DB
	if update {
		result = tx.Model(model).Where(where).Assign(data).FirstOrCreate(&data)
	} else {
		result = tx.Model(model).Where(where).Create(data)
	}
	if result.Error != nil {
		// tx.Rollback()

		return false, result.Error
	}
	// if DB == nil {
	// 	tx.Commit()
	// }
	return result.RowsAffected > 0, nil
}

func GetDependentEvents(event *entities.Event) (*[]entities.Event, error) {

	data := []entities.Event{}
	// err := db.SqlDb.Where(
	// 	&models.AuthorizationEvent{Event: entities.Event{PreviousEvent: *entities.NewEventPath(entities.AuthModel, event.Hash)}},
	// ).Or(&models.AuthorizationEvent{Event: entities.Event{AuthEvent: *entities.NewEventPath(entities.AuthModel, event.Hash)}},
	// // ).Or("? LIKE ANY (associations)", fmt.Sprintf("%%%s%%", event.Hash)
	// ).Find(&data).Error
	// if err != nil {
	// 	return nil, err
	// }
	prevEvent, _ := GetEventFromPath(&(event.PreviousEvent))
	if prevEvent != nil {
		data = append(data, *prevEvent)
	}
	authEvent, _ := GetEventFromPath(&(event.AuthEvent))
	if authEvent != nil {
		data = append(data, *authEvent)
	}
	return &data, nil
}

// func IncrementRecord[Model any](where Model, field string, DB *gorm.DB) (model *Model, created bool, err error) {
// 	tx := DB
// 	if DB == nil {
// 		tx = sql.SqlDb.Begin()
// 	}
// 	// dataByte, err := encoder.MsgPackStruct(event.Payload)
// 	if err != nil {
// 		return nil, false, err
// 	}
// 	// authPayload := entities.ClientPayload{
// 	// 	Data: entities.Authorization{},
// 	// }
// 	// copier.Copy(&authPayload, &event.Payload)
// 	// data := models.AuthorizationEvent{
// 	// 	Event: *event,
// 	// }
// 	var result *gorm.DB
// 	var data Model
// 	logger.Debugf("111111::::: %s", where)
// 	result = tx.Where(where).Find(data)
// 	if result.Error != nil {
// 		tx.Rollback()

// 		return nil, false, result.Error
// 	}
// 	if result.RowsAffected == 0 {
// 		logger.Debugf("111111 TTTT ::::: %s", where)
// 		result = tx.Where(where).FirstOrCreate(&data)
// 	} else {
// 		logger.Debugf("111111 FFFF ::::: %s", where)
// 		result = tx.Where(where).Update(field, fmt.Sprintf("%s + 1", field))

// 	}
// 	// if update {
// 	// 	result = tx.Where(where).Assign(data).FirstOrCreate(&data)
// 	// } else {
// 	// 	result = tx.Where(where).FirstOrCreate(&data)
// 	// }
// 	logger.Debugf("22222 ::::: %s", where)
// 	if result.Error != nil {
// 		tx.Rollback()

// 		return nil, false, result.Error
// 	}
// 	logger.Debugf("33333 ::::: %s", where)
// 	if DB == nil {
// 		if tx.Commit().Error != nil {
// 			tx.Rollback()
// 			return nil, false, tx.Error
// 		}

// 	}
// 	logger.Debugf("44444 ::::: %s", where)
// 	return &data, result.RowsAffected > 0, nil
// }

func GetTableName(table any) string {
	stmt := &gorm.Statement{DB: sql.SqlDb}
	stmt.Parse(table)
	return stmt.Schema.Table
}

func GetAccountSubscriptions(account string) {

	topicTableName := GetTableName(models.TopicState{})
	subscriptionTableName := GetTableName(models.SubscriptionState{})

	rows, err := sql.SqlDb.Table(subscriptionTableName).Where(fmt.Sprintf("%s.subscriber = \"%s\"", subscriptionTableName, account)).Joins(fmt.Sprintf("right join %s on %s.topic = %s.id", topicTableName, subscriptionTableName, topicTableName)).Rows()
	defer rows.Close()

	// logger.Debug(rows.)

	if err != nil {
		logger.Debug(err)
	}

	var subscriptions []map[string]string
	for rows.Next() {
		var subscription map[string]string
		sql.SqlDb.ScanRows(rows, &subscription)
		subscriptions = append(subscriptions, subscription)
	}
	logger.Debugf("%s", subscriptions)

}
func FormatSQL(vars []interface{}) string {
	query := ""
	for i:=0; i< len(vars); i++ {
		query = fmt.Sprintf("%s%s, ", query, "?")
	}
	query = query[0:strings.LastIndex(query, ",")]
	for _, v := range vars {
		// Convert each variable to a string and safely replace `?` with it
		if( reflect.ValueOf(v).Kind() == reflect.String || reflect.TypeOf(v) == reflect.TypeOf(time.Time{}))  {
			query = strings.Replace(query, "?", fmt.Sprintf("'%v'", v), 1)
		} else {
			if reflect.TypeOf(v) == nil {
				query = strings.Replace(query, "?", "NULL", 1)
			} else {
				if reflect.TypeOf(v) == reflect.TypeOf(dbsql.NullTime{}) {
					query = strings.Replace(query, "?", fmt.Sprintf("%v", v.(dbsql.NullTime).Time), 1)
				}
				if reflect.TypeOf(v) == reflect.TypeOf(dbsql.NullBool{}) {
					query = strings.Replace(query, "?", fmt.Sprintf("%v", v.(dbsql.NullBool).Bool), 1)
				}
				if reflect.TypeOf(v) == reflect.TypeOf(dbsql.NullFloat64{}) {
					query = strings.Replace(query, "?", fmt.Sprintf("%v", v.(dbsql.NullFloat64).Float64), 1)
				}
				query = strings.Replace(query, "?", fmt.Sprintf("%v", v), 1)
			}
		}
		
	}
	return query
}

type ExportData struct {
	Table string `json:"table"`
	Columns string `json:"cols"`
	Data [][]interface{} `json:"d"`
}
func GenerateImportScript[T any](db *gorm.DB, model T, where string, fileName string, cfg *configs.MainConfiguration ) ([]byte, error) {
    // var sqlScript string
	// var rows  []T	
	
	//result := sql.SqlDb.Table(GetTableName(model)).Where(where).Order("created_at DESC").Find(&rows)
    // if result.Error != nil {
    //     logger.Error(result.Error)
	// 	return nil, result.Error
    // }
	tableName := GetTableName(model)
	Db, err := sql.SqlDb.DB()
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(tableName, "_event") {
		where += " AND synced = true AND broadcasted = true"
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, where)
	
	
	rows, err := Db.Query(query)
    if err != nil {
        logger.Error(err)
		return nil, err
    }
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
        logger.Error(err)
		return nil, err
    }
	var results [][]interface{}
	for rows.Next() {
        // Create a slice of interfaces to hold values for the current row
        values := make([]interface{}, len(columns))
        valuePtrs := make([]interface{}, len(columns))

        for i := range values {
            valuePtrs[i] = &values[i]
        }

        // Scan the row into the value pointers
        if err := rows.Scan(valuePtrs...); err != nil {
            return nil, err
        }

        // Create a map for the current row
        // rowMap := make(map[string]interface{})
		var row []interface{}
        for i := 0;  i<len(columns); i++ {
            row = append(row, values[i])
        }
        // Append the map to results
        results = append(results, row)
    }
	
   
	data := ExportData{Table: tableName}
	data.Columns = strings.Join(columns, ",")
	data.Data = results
	s, err := encoder.MsgPackStruct(data)
	if err != nil {
		return nil, err
	}
// 	data := ExportData{Table: GetTableName(model)}
// 	if (len(rows) > 0) {
// 		stmt := db.Session(&gorm.Session{DryRun: true}).Clauses(clause.OnConflict{
//             DoNothing: true, // To generate INSERT OR IGNORE, use DoNothing
//         }).Model(model).Create(&rows[0]).Statement
		
//         sql := stmt.SQL.String()
// 		columns := sql[strings.Index(sql, "(")+1:]
// 		data.Columns = columns[0:strings.Index(columns,")")]
// 	}
	
	
//      for _, row := range rows {
// 		// rowsData, err :=  json.Marshal(row)
// 		// if err != nil {
// 		// 	logger.Fatal(err)
// 		// }
// 		// row2 := map[string]interface{}{}
// 		// err = json.Unmarshal(rowsData, &row2)
// 		// if err != nil{
// 		// 	logger.Fatal(err)
// 		// }
//         // Create a new DB session with DryRun mode
// 		utils.SetDefaultValues(&row)
		
// 		// columArray := strings.Split(data.Columns, ",")
//         stmt := db.Session(&gorm.Session{DryRun: true}).Clauses(clause.OnConflict{
//             DoNothing: true, // To generate INSERT OR IGNORE, use DoNothing
//         }).Model(model).Create(&row).Statement
// 		data.Data = append(data.Data, stmt.Vars)
// 		logger.Infof("STATMENT %s", stmt.SQL.String(),stmt.Statement.Vars)
// 		// rowByte, _ := json.Marshal(row)
// 		// rowMap := map[string]any{}
// 		// json.Unmarshal(rowByte, &rowMap)
// 		// _vars := []any{}
// 		// for _, col := range columArray {
// 		// 	logger.Infof("COLLLL", col, rowMap)
// 		// 	_vars = append(_vars, utils.GetFieldValueByName(row, col))
// 		// }
// 		logger.Infof("ROWLLEN %d, %d, %v", len(strings.Split(data.Columns, ",")), len(stmt.Statement.Vars))
		
//         // sql := stmt.SQL.String()
// 		// columns := sql[strings.Index(sql, "(")+1:]
// 		// columns = columns[0:strings.Index(columns,")")]
// 		// logger.Debug("COLUMNS",  columns)
// 		// sql = formatSQL(sql, stmt.Vars)
// 		// // 	logger.Debug("QUERY",  stmt.Vars)
// 		// 	// break
//         // // Replace "INSERT" with "INSERT OR REPLACE" in the generated SQL query
//         // sql = "INSERT OR REPLACE" + sql[len("INSERT"):]
		
//         // // Append the generated SQL query to the script
//         // sqlScript += sql + ";\n"
//    } 
//    s, err :=  json.Marshal(data)
//    if err != nil {
// 	return nil, err
//    }
// 	if fileName != "" {
// 		fileName = fmt.Sprintf("/tmp/%s", fileName)
// 		return s, SaveToFile(fileName, string(s))
// 	}
    return s, nil
}

func SaveToFile(filename, data string) (error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

	_, err = file.WriteString(data)
    if err != nil {
        return err
    }

    return nil
}

func ImportDataPostgres(cfg *configs.MainConfiguration, table string, columns []string, data [][]string) (error) {
	// if sql.Driver(cfg.SQLDB.DbDialect) != sql.Postgres {
	// 	return fmt.Errorf("invalid db")
	// }
	dir := filepath.Join(cfg.DataDir, "tmp")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return  err
	}
	header := [][]string{columns}
	csvFile := filepath.Join(dir,  utils.RandomAplhaNumString(6))
	err = utils.WriteToCSV(csvFile, header)
	if err != nil {
		return err
	}
	err = utils.WriteToCSV(csvFile, data)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("COPY %s FROM '%s' DELIMITER ',' CSV HEADER;", table, csvFile)
	result :=  sql.SqlDb.Exec(query)
	if result.Error == nil {
		os.Remove(csvFile)
	} 
	return result.Error
}

// func ImportDataSqlite(cfg *configs.MainConfiguration, tableName string, rows [][]interface{}) (error) {
// 	if sql.Driver(cfg.SQLDB.DbDialect) != sql.Sqlite {
// 		return fmt.Errorf("invalid db")
// 	}
// 	// err = sql.SqlDb.Transaction(func(tx *gorm.DB) error {
// 	// 	// Prepare SQL for bulk insert without using structs
// 	// 	valueStrings := make([]string, 0, len(rows))
// 	// 	valueArgs := make([]interface{}, 0, len(rows[0]))
// 	// 	values := []interface{}
// 	// 	for _, row := range rows {
// 	// 		values = append(values, FormatSQL(row))
// 	// 	}
// 	// 		if i==len(b.Data)-1 || i+1 % batchSize == 0 {
// 	// 			// save it
// 	// 			// sql.SqlDb.Exec()
// 	// 			query := fmt.Sprintf("INSERT INTO %s (%s) values %s", b.Table, b.Columns, values)
// 	// 			logger.Debug("QUERYYY", query)
// 	// 			values = ""
// 	// 		}
// 	// 	}

// 	// 	// Join the value placeholders and construct the final SQL query
// 	// 	insertSQL := fmt.Sprintf("INSERT INTO %s (name) VALUES %s", tableName, strings.Join(valueStrings, ","))

// 	// 	// Execute the raw SQL query with arguments inside the transaction
// 	// 	if err := tx.Exec(insertSQL, valueArgs...).Error; err != nil {
// 	// 		return err // Rollback the transaction if an error occurs
// 	// 	}

// 	// 	// Commit the transaction
// 	// 	return nil
// 	// })
// 	return result.Error
// }