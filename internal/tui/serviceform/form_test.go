package serviceform

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func press(m model, code rune, mod tea.KeyMod) model {
	next, _ := m.Update(tea.KeyPressMsg{Code: code, Mod: mod})
	return next.(model)
}

func setValue(m *model, id fieldID, value string) {
	for i := range m.fields {
		if m.fields[i].id == id {
			m.fields[i].input.SetValue(value)
		}
	}
}

func validModel() model {
	m := newModel("nas")
	setValue(&m, fqdnField, "nas.home")
	setValue(&m, ipField, "192.168.1.50")
	setValue(&m, portField, "443")
	return m
}

func TestCollapsedFieldsPreserveValuesAndSkipFocus(t *testing.T) {
	m := validModel()
	m.focus = advancedField
	m.syncFocus()
	m = press(m, tea.KeyEnter, 0)
	m = press(m, tea.KeyTab, 0)

	if m.focus != runbookField {
		t.Fatal("expanded field is not reachable")
	}

	next, _ := m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})

	m = next.(model)
	m = press(m, tea.KeyTab, tea.ModShift)
	m = press(m, tea.KeySpace, 0)
	m = press(m, tea.KeyTab, 0)

	if m.advancedOpen || m.focus != saveField || m.value(runbookField) != "a" || m.value(fqdnField) != "nas.home" {
		t.Fatal("collapse must preserve edits and skip hidden fields")
	}
}

func TestChecksAndSave(t *testing.T) {
	m := validModel()
	m.focus = checksField
	m.syncFocus()

	m = press(m, tea.KeySpace, 0)
	if m.checks[0].selected {
		t.Fatal("space did not toggle")
	}

	setValue(&m, runbookField, "runbooks/nas.md")
	setValue(&m, expectedField, "204")
	setValue(&m, timeoutField, "3s")

	m.focus = saveField
	m = press(m, tea.KeyEnter, 0)

	if !m.submitted || m.service.Runbook != "runbooks/nas.md" || m.service.ExpectedStatus != 204 || m.service.Timeout != 3*time.Second || len(m.service.Checks) != 3 {
		t.Fatalf("submitted=%v, service=%+v", m.submitted, m.service)
	}
}

func TestDefaultsAndCancel(t *testing.T) {
	m := validModel()

	service, failure := m.submit()
	if failure != nil || service.EffectiveTimeout() != 5*time.Second || service.EffectiveStatusCode() != 200 {
		t.Fatal("blank defaults failed")
	}

	m = press(m, tea.KeyEnter, 0)
	if m.submitted {
		t.Fatal("enter in text field submitted")
	}
	m = press(m, tea.KeyEsc, 0)
	if m.submitted {
		t.Fatal("cancel submitted")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var out bytes.Buffer
	service, err := Run(ctx, strings.NewReader(""), &out, "nas")
	if !errors.Is(err, context.Canceled) || service.FQDN != "" || out.Len() != 0 {
		t.Fatal("context cancellation ignored")
	}
}
