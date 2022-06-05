package privatevpn

import (
	"github.com/qdm12/gluetun/internal/configuration/settings"
	"github.com/qdm12/gluetun/internal/constants/providers"
	"github.com/qdm12/gluetun/internal/models"
	"github.com/qdm12/gluetun/internal/provider/utils"
)

func (p *Provider) GetConnection(selection settings.ServerSelection) (
	connection models.Connection, err error) {
	defaults := utils.NewConnectionDefaults(443, 1194, 0) //nolint:gomnd
	return utils.GetConnection(providers.Privatevpn,
		p.storage, selection, defaults, p.randSource)
}
