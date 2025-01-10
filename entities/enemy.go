package entities

import "rpg-sphere/components"

type Enemy struct {
	*Sprite
	FollowsPlayer bool
	CombatComp    *components.EnemyCombat
}
