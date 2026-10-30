package status

import "testing"

func TestRowsAreSortedAndCopied(t *testing.T) {
	s := &Store{}
	in := map[string]string{"b": "2", "a": "1", "%x%": "0", "Status": "ok"}
	s.Set(in)
	in["a"] = "changed"
	rows := s.Rows()
	if len(rows) != 4 || rows[0].Key != "Status" || rows[1].Key != "%x%" || rows[2].Key != "a" || rows[2].Value != "1" || rows[3].Key != "b" {
		t.Errorf("%v", rows)
	}
}

func TestEmptyStore(t *testing.T) {
	if rows := (&Store{}).Rows(); len(rows) != 0 {
		t.Errorf("%v", rows)
	}
}

func TestOnChangeSeesNewRows(t *testing.T) {
	s := &Store{}
	calls := 0
	s.OnChange(func() {
		calls++
		if rows := s.Rows(); len(rows) != 1 || rows[0].Value != "v" {
			t.Errorf("listener saw %v", rows)
		}
	})
	s.Set(map[string]string{"k": "v"})
	if calls != 1 {
		t.Errorf("%d calls", calls)
	}
	s.OnChange(nil)
	s.Set(nil)
}
