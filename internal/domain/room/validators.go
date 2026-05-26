package room

import "errors"

func (p *Patch) Validate() error {
	switch {
	case p.Action == ActionAddType && p.Content == nil: return errors.New("there must have the content in action add type")
	case p.Action == ActionDeleteType && p.EndIndex == nil: return errors.New("there must have end index in delete action type") 
	default: return nil
	}
}