package profiles

import (
	"fmt"

	"github.com/mikeunge/sshman/internal/database"

	"atomicgo.dev/keyboard/keys"
	"github.com/pterm/pterm"
)

func (s *ProfileService) multiSelectProfiles(t string, maxHeight int) ([]int64, error) {
	var profiles []database.SSHProfile
	var selectedProfiles []int64
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return selectedProfiles, err
	}

	var pProfiles []string
	for _, p := range profiles {
		authType := database.GetNameFromAuthType(p.AuthType)
		pProfiles = append(pProfiles, fmt.Sprintf("%d %s %s@%s (%s)", p.Id, p.Alias, p.User, p.Host, authType))
	}

	height := len(pProfiles)
	if len(pProfiles) > maxHeight && maxHeight > 0 {
		height = maxHeight
	}

	selectedOptions, _ := pterm.DefaultInteractiveMultiselect.
		WithDefaultText(t).
		WithOptions(pProfiles).
		WithMaxHeight(height).
		WithFilter(false).
		WithKeyConfirm(keys.Enter).
		WithKeySelect(keys.Space).
		WithCheckmark(&pterm.Checkmark{Checked: pterm.Green("+"), Unchecked: pterm.Red("-")}).
		Show()

	if selectedProfiles, err = parseIdsFromSelectedProfiles(selectedOptions); err != nil {
		return selectedProfiles, err
	}
	return selectedProfiles, nil
}

func (s *ProfileService) selectProfile(t string, maxHeight int) (int64, error) {
	var profiles []database.SSHProfile
	var err error

	if profiles, err = s.DB.GetAllSSHProfiles(); err != nil {
		return 0, err
	}

	var pProfiles []string
	for _, p := range profiles {
		authType := database.GetNameFromAuthType(p.AuthType)
		pProfiles = append(pProfiles, fmt.Sprintf("%d %s %s@%s (%s)", p.Id, p.Alias, p.User, p.Host, authType))
	}

	height := len(pProfiles)
	if len(pProfiles) > maxHeight && maxHeight > 0 {
		height = maxHeight
	}

	selectedOptions, err := pterm.DefaultInteractiveSelect.
		WithDefaultText(t).
		WithOptions(pProfiles).
		WithMaxHeight(height).
		WithFilter(false).
		Show()

	if err != nil {
		return 0, err
	}

	var parsedProfileIds []int64
	if parsedProfileIds, err = parseIdsFromSelectedProfiles([]string{selectedOptions}); err != nil {
		return 0, err
	}
	if len(parsedProfileIds) == 0 {
		return 0, fmt.Errorf("could not parse id")
	}
	return parsedProfileIds[0], nil
}

func prettyPrintProfiles(profiles []database.SSHProfile) {
	var data [][]string
	var dFormat = "02.01.2006"

	data = append(data, []string{"Id", "Alias", "User", "Host/IP", "Authentication", "Encrypted", "Created At"}) // define the table header
	for _, profile := range profiles {
		encrypted := "-"
		authType := database.GetNameFromAuthType(profile.AuthType)
		if profile.Encrypted {
			encrypted = "+"
		}
		data = append(data, []string{fmt.Sprintf("%d", profile.Id), profile.Alias, profile.User, profile.Host, authType, encrypted, profile.CTime.Format(dFormat)})
	}
	pterm.DefaultTable.
		WithHasHeader().
		WithData(data).
		Render()
}
