package adventure

import (
	"fisherevans.com/project/f/internal/resources"
)

type Zones map[string]struct{}

func (z Zones) Contains(a string) bool {
	if z == nil {
		return false
	}
	_, contains := z[a]
	return contains
}

type zones struct {
	zones map[MapLocation]map[string]struct{}
}

func newZones() *zones {
	return &zones{
		zones: make(map[MapLocation]map[string]struct{}),
	}
}

func (z *zones) SetZoneId(loc MapLocation, zoneId string) {
	if _, exists := z.zones[loc]; !exists {
		z.zones[loc] = make(map[string]struct{})
	}
	z.zones[loc][zoneId] = struct{}{}
}

func (z *zones) RegisterZone(zone resources.Zone) {
	for x := zone.X; x <= zone.X+zone.W; x++ {
		for y := zone.Y; y <= zone.Y+zone.H; y++ {
			loc := MapLocation{X: x, Y: y}
			if _, exists := z.zones[loc]; !exists {
				z.zones[loc] = make(map[string]struct{})
			}
			z.zones[loc][zone.ZoneId] = struct{}{}
		}
	}
}

func (z *zones) ZonesAt(lol MapLocation) []string {
	zoneSet, hasAny := z.zones[lol]
	if !hasAny {
		return nil
	}
	var zoneList []string
	for zoneId := range zoneSet {
		zoneList = append(zoneList, zoneId)
	}
	return zoneList
}

func (z *zones) ZonesAtSet(lol MapLocation) Zones {
	zoneSet, hasAny := z.zones[lol]
	if !hasAny {
		return map[string]struct{}{}
	}
	return zoneSet
}
