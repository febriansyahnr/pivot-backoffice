package xlsx

import (
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/xuri/excelize/v2"
)

const (
	CustomNormalCurrencyFmt          = "[$Rp -421]#,##0"
	CustomNormalCurrencyWithColorFmt = "+ [$Rp -421]#,##0;- [$Rp -421]#,##0"
	CustomDatetimefmt                = "dd mmm yyyy, hh:mm AM/PM"
)

func DefaultRowOpt() excelize.RowOpts {
	return excelize.RowOpts{
		Height: 18,
	}
}

func StyleTitle(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 12,
		},
	})
	return s
}

func StyleTitleWithAlignment(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 12,
		},
		Alignment: &excelize.Alignment{
			Vertical: "center", Horizontal: "left",
		},
	})
	return s
}

func StyleSubTitle(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 10,
		},
	})
	return s
}

func StyleHeaderColumn(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true, Family: "Arial", Size: 10,
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill: excelize.Fill{
			Type: "pattern", Pattern: 1, Color: []string{"#e4e4e4"},
		},
	})
	return s
}

func StyleNormalRow(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
	})
	return s
}

func StyleCurrencyRow(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
		CustomNumFmt: util.ValueToPtr(CustomNormalCurrencyFmt),
	})
	return s
}

func StyleCurrencyRowWithCustomFormat(f *File, customFmt string) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
		CustomNumFmt: util.ValueToPtr(customFmt),
	})
	return s
}

func StyleDatetime12HoursRow(f *File) int {
	s, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: false, Family: "Arial", Size: 10,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
		},
		CustomNumFmt: util.ValueToPtr(CustomDatetimefmt),
	})
	return s
}
