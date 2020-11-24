package router

import (
	"fmt"
	"github.com/UpCloudLtd/cli/internal/commands"
	"github.com/UpCloudLtd/cli/internal/commands/network"
	"github.com/UpCloudLtd/cli/internal/ui"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
	"github.com/jedib0t/go-pretty/v6/table"
	"io"
	"sync"
)

func ShowCommand(service *service.Service) commands.Command {
	return &showCommand{
		BaseCommand: commands.New("show", "Show current router"),
		service:     service,
	}
}

type showCommand struct {
	*commands.BaseCommand
	service *service.Service
}

type routerWithNetworks struct {
	router   *upcloud.Router
	networks []*upcloud.Network
}

func (s *showCommand) MakeExecuteCommand() func(args []string) (interface{}, error) {
	return func(args []string) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("one router uuid or name is required")
		}
		r, err := searchRouter(args[0], s.service)
		if err != nil {
			return nil, err
		}

		var networks []*upcloud.Network
		var wg sync.WaitGroup
		var getNetworkError error

		for _, n := range r.AttachedNetworks {
			wg.Add(1)
			go func() {
				defer wg.Done()
				nw, err := network.SearchNetwork(n.NetworkUUID, s.service)
				if err != nil {
					getNetworkError = err
				}
				networks = append(networks, nw)
			}()
		}
		wg.Wait()
		if getNetworkError != nil {
			return nil, getNetworkError
		}
		return &routerWithNetworks{
			router:   r,
			networks: networks,
		}, nil
	}
}

func (s *showCommand) HandleOutput(writer io.Writer, out interface{}) error {
	routerWithNetworks := out.(*routerWithNetworks)
	r := routerWithNetworks.router
	networks := routerWithNetworks.networks

	l := ui.NewListLayout(ui.ListLayoutDefault)

	dCommon := ui.NewDetailsView()
	dCommon.AppendRows([]table.Row{
		{"UUID:", ui.DefaultUuidColours.Sprint(r.UUID)},
		{"Name:", r.Name},
		{"Type:", r.Type},
	})
	l.AppendSection("Common", dCommon.Render())

	tIPRouter := ui.NewDataTable("UUID", "Name", "Router", "Type", "Zone")
	for _, n := range networks {
		tIPRouter.AppendRow(table.Row{
			ui.DefaultUuidColours.Sprint(n.UUID),
			n.Name,
			n.Router,
			n.Type,
			n.Zone,
		})
	}
	l.AppendSection("Networks:", tIPRouter.Render())

	_, _ = fmt.Fprintln(writer, l.Render())
	return nil
}