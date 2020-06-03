// Code generated by protoc-gen-fieldmask. DO NOT EDIT.

package ttnpb

import (
	fmt "fmt"
	time "time"
)

func (dst *Event) SetFields(src *Event, paths ...string) error {
	for name, subs := range _processPaths(paths) {
		switch name {
		case "name":
			if len(subs) > 0 {
				return fmt.Errorf("'name' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Name = src.Name
			} else {
				var zero string
				dst.Name = zero
			}
		case "time":
			if len(subs) > 0 {
				return fmt.Errorf("'time' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Time = src.Time
			} else {
				var zero time.Time
				dst.Time = zero
			}
		case "identifiers":
			if len(subs) > 0 {
				return fmt.Errorf("'identifiers' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Identifiers = src.Identifiers
			} else {
				dst.Identifiers = nil
			}
		case "data":
			if len(subs) > 0 {
				return fmt.Errorf("'data' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Data = src.Data
			} else {
				dst.Data = nil
			}
		case "correlation_ids":
			if len(subs) > 0 {
				return fmt.Errorf("'correlation_ids' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.CorrelationIDs = src.CorrelationIDs
			} else {
				dst.CorrelationIDs = nil
			}
		case "origin":
			if len(subs) > 0 {
				return fmt.Errorf("'origin' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Origin = src.Origin
			} else {
				var zero string
				dst.Origin = zero
			}
		case "context":
			if len(subs) > 0 {
				return fmt.Errorf("'context' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Context = src.Context
			} else {
				dst.Context = nil
			}
		case "visibility":
			if len(subs) > 0 {
				var newDst, newSrc *Rights
				if (src == nil || src.Visibility == nil) && dst.Visibility == nil {
					continue
				}
				if src != nil {
					newSrc = src.Visibility
				}
				if dst.Visibility != nil {
					newDst = dst.Visibility
				} else {
					newDst = &Rights{}
					dst.Visibility = newDst
				}
				if err := newDst.SetFields(newSrc, subs...); err != nil {
					return err
				}
			} else {
				if src != nil {
					dst.Visibility = src.Visibility
				} else {
					dst.Visibility = nil
				}
			}
		case "authentication":
			if len(subs) > 0 {
				var newDst, newSrc *Event_Authentication
				if (src == nil || src.Authentication == nil) && dst.Authentication == nil {
					continue
				}
				if src != nil {
					newSrc = src.Authentication
				}
				if dst.Authentication != nil {
					newDst = dst.Authentication
				} else {
					newDst = &Event_Authentication{}
					dst.Authentication = newDst
				}
				if err := newDst.SetFields(newSrc, subs...); err != nil {
					return err
				}
			} else {
				if src != nil {
					dst.Authentication = src.Authentication
				} else {
					dst.Authentication = nil
				}
			}

		default:
			return fmt.Errorf("invalid field: '%s'", name)
		}
	}
	return nil
}

func (dst *StreamEventsRequest) SetFields(src *StreamEventsRequest, paths ...string) error {
	for name, subs := range _processPaths(paths) {
		switch name {
		case "identifiers":
			if len(subs) > 0 {
				return fmt.Errorf("'identifiers' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Identifiers = src.Identifiers
			} else {
				dst.Identifiers = nil
			}
		case "tail":
			if len(subs) > 0 {
				return fmt.Errorf("'tail' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Tail = src.Tail
			} else {
				var zero uint32
				dst.Tail = zero
			}
		case "after":
			if len(subs) > 0 {
				return fmt.Errorf("'after' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.After = src.After
			} else {
				dst.After = nil
			}

		default:
			return fmt.Errorf("invalid field: '%s'", name)
		}
	}
	return nil
}

func (dst *Event_Authentication) SetFields(src *Event_Authentication, paths ...string) error {
	for name, subs := range _processPaths(paths) {
		switch name {
		case "type":
			if len(subs) > 0 {
				return fmt.Errorf("'type' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.Type = src.Type
			} else {
				var zero string
				dst.Type = zero
			}
		case "token_type":
			if len(subs) > 0 {
				return fmt.Errorf("'token_type' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.TokenType = src.TokenType
			} else {
				var zero string
				dst.TokenType = zero
			}
		case "token_id":
			if len(subs) > 0 {
				return fmt.Errorf("'token_id' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.TokenID = src.TokenID
			} else {
				var zero string
				dst.TokenID = zero
			}
		case "remote_ip":
			if len(subs) > 0 {
				return fmt.Errorf("'remote_ip' has no subfields, but %s were specified", subs)
			}
			if src != nil {
				dst.RemoteIP = src.RemoteIP
			} else {
				var zero string
				dst.RemoteIP = zero
			}

		default:
			return fmt.Errorf("invalid field: '%s'", name)
		}
	}
	return nil
}
