package sql

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mlayerprotocol/go-mlayer/configs"
	config "github.com/mlayerprotocol/go-mlayer/configs"
	"github.com/mlayerprotocol/go-mlayer/internal/sql/models"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql/migration"
	"github.com/mlayerprotocol/go-mlayer/pkg/core/sql/sqlite"
	"github.com/mlayerprotocol/go-mlayer/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	dbLogger "gorm.io/gorm/logger"
)

var SqlDb *gorm.DB
var SqlDBErr error

var logger = &log.Logger
type Driver string

const (
	Postgres Driver = "postgres"
	MySQL Driver = "mysql"
	Sqlite Driver = "sqlite"
)

func InitializeDb(cfg *configs.MainConfiguration) (*gorm.DB, error) {
	
	driver := Driver(cfg.SQLDB.DbDialect)
	dsn := getDSN(cfg)
	logger.Debugf("\nInitializing %s db", driver)
	var dialect gorm.Dialector
	switch driver {
	case Postgres:
		dialect = postgres.Open(dsn)
	case MySQL:
		dialect = mysql.Open(dsn)
	default:
		// dialect = sqlite.Open(dsn)
		dialect = sqlite.NewSQLiteDialector(dsn)
	}
	SqlDb, err := gorm.Open(dialect, &gorm.Config{
		Logger: dbLogger.Default.LogMode(logLevel(cfg)),
	})

	if err != nil {
		logger.Fatal(err)
	}
	
	if driver == "sqlite" {
		//d, _ := SqlDb.DB()
		SqlDb.Exec("PRAGMA busy_timeout = 5000")
	}
	
	
	return SqlDb, err
}

func GetTableName(table any, db *gorm.DB) string {
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(table)
	return stmt.Schema.Table
}

func Init(cfg *configs.MainConfiguration) {
	
	SqlDb, SqlDBErr = InitializeDb(cfg)
	if SqlDBErr != nil {
		panic(SqlDBErr)
	}

	for _, migration := range migration.Migrations {
		err := migration.Migrate(SqlDb)
			
			if err != nil {
				log.Logger.Error("Migration Error", err)
				panic(err)
			}
		// var m models.MigrationState;
		// key := strings.ToLower(fmt.Sprintf("%s:%s", migration.DateTime,  migration.Id))
		// err := SqlDb.Where(models.MigrationState{Key: key }).First(&m).Error
		// if err == gorm.ErrRecordNotFound {
		// 	err := migration.Migrate(SqlDb)
			
		// 	if err == nil {
		// 		SqlDb.Create(&models.MigrationState{Key: key })
		// 	} else {
		// 		log.Logger.Error("Migration Error", err)
		// 		panic(err)
		// 	}
		// }
	}
	for _, model := range models.Models {
		err := SqlDb.AutoMigrate(&model)
		if err != nil {
			logger.Errorf("SQL_MIGRATION_ERROR: %v", err)
		}
	}
	
	db, err := SqlDb.DB()
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(cfg.SQLDB.DbMaxConnLifetime)
	db.SetMaxOpenConns(cfg.SQLDB.DbMaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.SQLDB.DbMaxConnLifetime) * time.Second)
	// SqlDb.Exec("DROP TRIGGER IF EXISTS app_events_sync_trigger;")
	counterTable := GetTableName(models.EventCounter{}, SqlDb)
	appSyncTrigger, appSyncFunc := EventSyncedTrigger(config.Config.SQLDB.DbDialect, GetTableName(models.ApplicationEvent{}, SqlDb), counterTable)
	SqlDb.Exec(string(appSyncFunc))
	SqlDb.Exec(string(appSyncTrigger))

	authSyncTrigger, authSyncFunc := EventSyncedTrigger(config.Config.SQLDB.DbDialect, GetTableName(models.AuthorizationEvent{}, SqlDb), counterTable)
	SqlDb.Exec(string(authSyncFunc))
	SqlDb.Exec(string(authSyncTrigger))
	
	
}

func logLevel(cfg *configs.MainConfiguration) dbLogger.LogLevel {
	if cfg.LogLevel == "info" {
		return dbLogger.Warn
	}
	if cfg.LogLevel == "debug" {
		return dbLogger.Info
	}
	if strings.Contains(cfg.LogLevel, "warn")  {
		return dbLogger.Warn
	}
	return dbLogger.Warn
}

func getDSN(cfg *configs.MainConfiguration) string {
	dsn := ""
	switch strings.ToLower(config.Config.SQLDB.DbDialect) {
	case "sqlite":
		err := os.MkdirAll(cfg.SQLDB.DbStoragePath, os.ModePerm)
		if err != nil {
			logger.Errorf("Error creating sqlite storage directory at %s", config.Config.SQLDB.DbStoragePath)
			panic(err)
		}
		if strings.HasSuffix(cfg.SQLDB.DbStoragePath, "/") {
			dsn = fmt.Sprintf("%sdb.sqlite", cfg.SQLDB.DbStoragePath)
		} else {
			dsn = fmt.Sprintf("%s/db.sqlite", cfg.SQLDB.DbStoragePath)
		}
		logger.Debug("DBPATH",cfg.SQLDB.DbStoragePath )
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Config.SQLDB.DbUser, config.Config.SQLDB.DbPassword, config.Config.SQLDB.DbHost, config.Config.SQLDB.DbPort, config.Config.SQLDB.DbDatabase)
	// case "postgres":
	default:
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s", config.Config.SQLDB.DbHost, config.Config.SQLDB.DbUser, config.Config.SQLDB.DbPassword, config.Config.SQLDB.DbDatabase, config.Config.SQLDB.DbPort, config.Config.SQLDB.DbSSLMode, config.Config.SQLDB.DbTimezone)
	}
	return dsn
}
