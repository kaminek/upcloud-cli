package storage

import (
	// "sync"

	"github.com/UpCloudLtd/cli/internal/commands"
	"github.com/UpCloudLtd/cli/internal/output"
	"github.com/UpCloudLtd/cli/internal/resolver"
	"github.com/UpCloudLtd/cli/internal/ui"

	// "github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
)

// ShowCommand creates the "storage show" command
func ShowCommand() commands.Command {
	return &showCommand{
		BaseCommand: commands.New("show", "Show storage details"),
	}
}

type showCommand struct {
	*commands.BaseCommand
	resolver.CachingStorage
}

// InitCommand implements Command.InitCommand
func (s *showCommand) InitCommand() {
	// TODO: reimplmement
	// s.SetPositionalArgHelp(positionalArgHelp)
	// TODO: reimplmement
	// s.ArgCompletion(getStorageArgumentCompletionFunction(s.Config()))
}

// Execute implements Command.MakeExecuteCommand
func (s *showCommand) Execute(exec commands.Executor, uuid string) (output.Output, error) {
	// var (
	// wg sync.WaitGroup
	// storageImportDetailsErr error
	// )

	storageSvc := exec.Storage()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()

	// 	storageImport, storageImportDetailsErr = storageSvc.GetStorageImportDetails(
	// 		&request.GetStorageImportDetailsRequest{UUID: uuid},
	// 	)

	// 	if ucErr, ok := storageImportDetailsErr.(*upcloud.Error); ok {
	// 		if ucErr.ErrorCode == "STORAGE_IMPORT_NOT_FOUND" {
	// 			storageImportDetailsErr = nil
	// 		}
	// 	}
	// }()

	storage, err := storageSvc.GetStorageDetails(
		&request.GetStorageDetailsRequest{UUID: uuid},
	)
	if err != nil {
		return nil, err
	}

	// wg.Wait()
	// if storageImportDetailsErr != nil {
	// 	return nil, storageImportDetailsErr
	// }

	// Storage details
	storageSection := output.CombinedSection{
		Contents: output.Details{
			Sections: []output.DetailSection{
				{
					Title: "Storage",
					Rows: []output.DetailRow{
						{Title: "UUID:", Key: "uuid", Value: storage.UUID, Color: ui.DefaultUUUIDColours},
						{Title: "Title:", Key: "title", Value: storage.Title},
						{Title: "type:", Key: "type", Value: storage.Type},
						{Title: "State:", Key: "state", Value: storage.State, Color: commands.StorageStateColor(storage.State)},
						{Title: "Size:", Key: "size", Value: storage.Size},
						{Title: "Tier:", Key: "tier", Value: storage.Tier},
						{Title: "Zone:", Key: "zone", Value: storage.Zone},
						{Title: "Server:", Key: "zone", Value: storage.ServerUUIDs[0]},
						{Title: "Origin:", Key: "origin", Value: storage.Origin, Color: ui.DefaultUUUIDColours},
						{Title: "Created:", Key: "created", Value: storage.Created},
						{Title: "Licence:", Key: "licence", Value: storage.License},
					},
				},
			},
		},
	}

	combined := output.Combined{
		storageSection,
	}

	// Backups
	if storage.BackupRule != nil && storage.BackupRule.Interval != "" {
		combined = append(combined, output.CombinedSection{
			Contents: output.Details{
				Sections: []output.DetailSection{
					{
						Title: "Backup Rule",
						Rows: []output.DetailRow{
							{Title: "Interval:", Key: "interval", Value: storage.BackupRule.Interval},
							{Title: "Time:", Key: "time", Value: storage.BackupRule.Time},
							{Title: "Retention:", Key: "retention", Value: storage.BackupRule.Retention},
						},
					},
				},
			},
		})
	}

	if len(storage.BackupUUIDs) > 0 {
		backupsListRows := []output.TableRow{}
		for _, b := range storage.BackupUUIDs {
			backupsListRows = append(backupsListRows, output.TableRow{b})
		}
		combined = append(combined, output.CombinedSection{
			Key:   "available_backups",
			Title: "Available Backups",
			Contents: output.Table{
				Columns: []output.TableColumn{
					{Key: "uuid", Header: "UUID", Color: ui.DefaultUUUIDColours},
				},
				Rows: backupsListRows,
			},
		})
	}

	return combined, nil
}
