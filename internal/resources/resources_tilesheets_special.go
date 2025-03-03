package resources

func newTilesheetSpriteId(tilesheet string, col, row int) TilesheetSpriteId {
	return TilesheetSpriteId{
		Tilesheet: tilesheet,
		Column:    col,
		Row:       row,
	}
}

var (
	TileCollisionBlock          = newTilesheetSpriteId("snowhex_base", 65, 49)
	TileCollisionJumpVertical   = newTilesheetSpriteId("snowhex_base", 66, 49)
	TileCollisionJumpAll        = newTilesheetSpriteId("snowhex_base", 67, 49)
	TileCollisionJumpHorizontal = newTilesheetSpriteId("snowhex_base", 68, 49)
	TileCollisionStairsDown     = newTilesheetSpriteId("snowhex_base", 69, 49)
	TileCollisionStairsUp       = newTilesheetSpriteId("snowhex_base", 70, 49)
	TileCollisionDoor           = newTilesheetSpriteId("snowhex_base", 71, 49)

	SpriteButtonA            = newTilesheetSpriteId("ab_button_icons", 1, 1)
	SpriteButtonB            = newTilesheetSpriteId("ab_button_icons", 2, 1)
	SpriteButtonAHighlighted = newTilesheetSpriteId("ab_button_icons", 3, 1)
	SpriteButtonBHighlighted = newTilesheetSpriteId("ab_button_icons", 4, 1)
)
