package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
)

const (
    ANSI_CLEAR string = "\033[0m"
    ANSI_RED string  = "\033[31m"
    ANSI_GREEN string  = "\033[32m"
    ColumnKeyName string = "name"
    ColumnKeyFrom string = "from"
    ColumnKeyTo string = "to"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).
    Background(lipgloss.Color("57")).
    Bold(false)

var fromStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
var toStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

type Model struct {
	table table.Model
}

func newModel(updateString string) Model {
    columns := []table.Column{
        table.NewColumn(ColumnKeyName, "Name", 16).WithStyle(baseStyle),
        table.NewColumn(ColumnKeyFrom, "From", 16).WithStyle(baseStyle),
        table.NewColumn(ColumnKeyTo, "To", 16).WithStyle(baseStyle),
    }
    
    var rows []table.Row
    lines := strings.Split(updateString, "\n")
    for _, line := range lines {
        if len(line) < 4 {
            break 
        }
        subStrings := strings.Split(line, " ")
        rows = append(rows, table.NewRow(table.RowData{
            ColumnKeyName: subStrings[0],
            ColumnKeyFrom: table.NewStyledCell(subStrings[1], fromStyle),
            ColumnKeyTo: table.NewStyledCell(subStrings[3], toStyle),
        }))
    }
    keys := table.DefaultKeyMap()
    model := Model{
        table: table.New(columns).WithRows(rows).
        SelectableRows(true).
        WithBaseStyle(baseStyle).
        WithSelectedText(" ", "✓").
        SortByAsc(ColumnKeyName).
        WithKeyMap(keys).
        WithFooterVisibility(true).
        WithStaticFooter("lol").
        Focused(true),
    }
    
    model.UpdateFooter("normal")
    return model

} 

func (m Model) Init() tea.Cmd { return nil }

func (m Model) UpdateFooter(status string) {
    var footerText string
    //fmt.Println(status)
    switch status {
    case "normal":
        footerText = "[Space, ENTER] - Select package [U] - Update selected [A] - Update all"
    case "update":
        footerText = "Update in progress..."
    case "success":
        footerText = "Packages were successfully updates!"
    case "fail":
        footerText = "Failed to update packages :("
    }

    
    m.table = m.table.WithStaticFooter(footerText)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.UpdateFooter("normal")
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "u":
            
            shellCmd := exec.Command("run0", "pacman", "-Syu", "--noconfirm")
            m.UpdateFooter("update")
            output, err := shellCmd.CombinedOutput()
            if err != nil {
                fmt.Println("Error: ", err)
                fmt.Println("Output: ", output)
                m.UpdateFooter("fail")
                time.Sleep(5*time.Second)
                os.Exit(1)
            }
        }
    }
    m.table, cmd = m.table.Update(msg)
    return m, cmd
} 

func intInSlice(a int, list []int) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func (m Model) View() string {
    body := strings.Builder{}
    /*
    selectedIDs := []string{}
    for _, row := range m.table.SelectedRows() {
        selectedIDs = append(selectedIDs, row.Data[ColumnKeyName].(string))
    }
    */
    body.WriteString(m.table.View())
    body.WriteString("\n")
    body.WriteString("[U]- Update packages [R] - Refresh")
    return body.String()
}



func main () {
    cmd := exec.Command("checkupdates")
    var out bytes.Buffer
    cmd.Stdout = &out
    if err := cmd.Run(); err != nil {
        fmt.Println("All packages are up to date!")
        os.Exit(0)
    }
    updates := out.String()
    
    model := newModel(updates)

    if _, err := tea.NewProgram(model).Run(); err != nil {
        fmt.Println("Error running program", err)
        os.Exit(1)
    }
}


