package adventure

import "fmt"

type EntityCollision struct {
	InnateEntity
	Passable
}

func newCollisionEntity(location MapLocation) *EntityCollision {
	return &EntityCollision{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				id: EntityId(fmt.Sprintf("collision-%d-%d", location.X, location.Y)),
			},
			MapLocation: location,
		},
		Passable: newPassablePreventIngress(true),
	}
}

type EntityDashGap struct {
	InnateEntity
	Passable
}

func newDashGapEntity(location MapLocation) *EntityDashGap {
	return &EntityDashGap{
		InnateEntity: InnateEntity{
			BaseEntity: BaseEntity{
				id:             EntityId(fmt.Sprintf("dashgap-%d-%d", location.X, location.Y)),
				isInteractable: true,
			},
			MapLocation: location,
		},
		Passable: newPassablePreventIngress(true),
	}
}
