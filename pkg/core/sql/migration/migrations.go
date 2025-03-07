package migration

import (
	"reflect"
	"runtime"

	"gorm.io/gorm"
)

type Migration struct {
	Id       string
	DateTime string
	Migrate  func(db *gorm.DB) error
}

var Migrations = []Migration{}

func AddMigration(migration func(db *gorm.DB) error, dateTime string) {
	migrationName := runtime.FuncForPC(reflect.ValueOf(migration).Pointer()).Name()
	Migrations = append(Migrations, Migration{
		Id: migrationName,
		DateTime: dateTime,
		Migrate: migration,
	})
}

func init() {
	// Migrations = append(Migrations, Migration{
	// 	Id: "migrate-auth-index",
	// 	DateTime: "2024-04-15 6:00AM",
	// 	Migrate: MigrateAuthIndex,
	// })
	Migrations = append(Migrations, Migration{
		Id: "migrate-add-claimed-to-event-counter",
		DateTime: "2024-08-12 6:00AM",
		Migrate: AddClaimedFieldToEventCount,
	})

	AddMigration(DropOwnerColumnFromApplicationState, "2024-09-16 5:43PM") 
	AddMigration(DropTopicIdColumnFromMessageState, "2024-09-16 5:12PM") 
	AddMigration(DropAttachmentsColumnFromMessageState, "2024-09-16 5:42PM") 
	AddMigration(RenameEventAndAuthHashColumns, "2024-10-03 5:42PM") 
	// AddMigration(DropAgentColumnFromApplicationState, "2024-09-16=7 10:42AM") 

}


