package tasks

import "testing"

func TestParseInlineTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantText string
		wantType TaskType
		wantEff  TaskEffort
		wantLoc  TaskLocation
	}{
		{"empty string", "", "", "", "", ""},
		{"no tags", "buy milk", "buy milk", "", "", ""},
		{"type only", "#technical", "", TypeTechnical, "", ""},
		{"type at end", "buy milk #technical", "buy milk", TypeTechnical, "", ""},
		{"type and effort", "#technical buy milk @quick-win", "buy milk", TypeTechnical, EffortQuickWin, ""},
		{"tag mid-text", "buy #technical milk", "buy milk", TypeTechnical, "", ""},
		{"whitespace handling", "  #technical  buy milk  @quick-win  ", "buy milk", TypeTechnical, EffortQuickWin, ""},
		{"duplicate type last wins", "#technical #creative", "", TypeCreative, "", ""},
		{"case insensitive", "#TECHNICAL", "", TypeTechnical, "", ""},
		{"unrecognized token left in text", "#invalid task text", "#invalid task text", "", "", ""},
		{"all three tags", "buy milk #technical @quick-win +work", "buy milk", TypeTechnical, EffortQuickWin, LocationWork},
		{"all tag types", "#creative @deep-work +home", "", TypeCreative, EffortDeepWork, LocationHome},
		{"mixed case tags", "#Technical @Quick-Win +Work", "", TypeTechnical, EffortQuickWin, LocationWork},
		{"effort only", "@medium", "", "", EffortMedium, ""},
		{"location only", "+errands", "", "", "", LocationErrands},
		{"administrative type", "#administrative", "", TypeAdministrative, "", ""},
		{"physical type", "#physical", "", TypePhysical, "", ""},
		{"anywhere location", "+anywhere", "", "", "", LocationAnywhere},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotType, gotEff, gotLoc := ParseInlineTags(tt.input)
			if gotText != tt.wantText {
				t.Errorf("text = %q, want %q", gotText, tt.wantText)
			}
			if gotType != tt.wantType {
				t.Errorf("type = %q, want %q", gotType, tt.wantType)
			}
			if gotEff != tt.wantEff {
				t.Errorf("effort = %q, want %q", gotEff, tt.wantEff)
			}
			if gotLoc != tt.wantLoc {
				t.Errorf("location = %q, want %q", gotLoc, tt.wantLoc)
			}
		})
	}
}
