package sql

import "fmt"



type Trigger string
type TriggerFunction string

var Triggers = map[string]Trigger{}


func EventSyncedTrigger(dbDriver string, table string, counterTable string) (Trigger, TriggerFunction) {
	switch dbDriver {
	case "sqlite":
		return Trigger(
			fmt.Sprintf(`
			CREATE TRIGGER IF NOT EXISTS %s_sync_trigger
			AFTER UPDATE ON %s
			FOR EACH ROW
			WHEN  NEW.synced = 1
			BEGIN
				UPDATE %s
				SET count = count + 1
				WHERE cycle = NEW.cycle AND app = NEW.app AND validator = NEW.validator;

				-- If no row is updated, insert a new row
				INSERT INTO %s (app, cycle, count, validator)
				SELECT NEW.app, NEW.cycle, 1, NEW.validator
				WHERE (SELECT changes() = 0);
			END;
			`,  table, table, counterTable, counterTable),
		), TriggerFunction("")
	case "postgres":
		return Trigger(fmt.Sprintf(`
		CREATE TRIGGER IF NOT EXISTS %s_sync_trigger
		AFTER UPDATE ON %s
		FOR EACH ROW
		WHEN NEW.synced = 1
		FOR EACH ROW
		WHEN (OLD.synced IS DISTINCT FROM NEW.synced)
		EXECUTE FUNCTION %s_sync_func();
		END;
	`, table, table, table)),
	TriggerFunction(fmt.Sprintf(`
		CREATE OR REPLACE FUNCTION %s_sync_func() RETURNS TRIGGER AS $$
    		BEGIN
				IF NEW.synced = 1 THEN
					UPDATE %s
					SET count = count + 1
					WHERE cycle = NEW.cycle AND app = NEW.app AND validator = NEW.validator;

					IF NOT FOUND THEN
						INSERT INTO %s (app, cycle, count, validator)
						VALUES (NEW.app, NEW.cycle, 1, NEW.validator)
						ON CONFLICT (app, cycle, validator) DO UPDATE
						SET count = %s.count + 1;
					END IF;
				END IF;
				RETURN NEW;
			END;
    	`, table, counterTable, counterTable, counterTable))
	
	}
	return "", ""
}