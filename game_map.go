package main

import (
	"github.com/faiface/pixel"
	"github.com/salviati/go-tmx/tmx"
)

type GameMap struct {
	mX, mY int

	// To track the camera position
	mCamX, mCamY float64

	mTilemap *tmx.Map
	mSprites map[string]*pixel.Sprite

	mTileSprite     pixel.Sprite
	mWidth, mHeight int

	mTiles        []*pixel.Batch
	mTilesIndices map[string]int
	mTilesCounter int

	mTileWidth, mTileHeight int
}

func (m *GameMap) Create(tilemap *tmx.Map) {
	// assuming exported tiled map
	//lua definition has 1 layer
	m.mTilemap = tilemap

	m.mHeight = tilemap.Height
	m.mWidth = tilemap.Width

	m.mTileWidth = tilemap.TileWidth
	m.mTileHeight = tilemap.TileHeight

	//Bottom left corner of the map, since pixel starts at 0, 0
	m.mX = m.mTileWidth
	m.mY = m.mTileHeight

	m.SetTiles()
}

func (m *GameMap) SetTiles() {
	batches := make([]*pixel.Batch, 0)
	batchIndices := make(map[string]int)
	batchCounter := 0

	// Load the sprites
	sprites := make(map[string]*pixel.Sprite)
	for _, tileset := range m.mTilemap.Tilesets {
		if _, alreadyLoaded := sprites[tileset.Image.Source]; !alreadyLoaded {
			sprite, pictureData := LoadSprite(tileset.Image.Source)
			sprites[tileset.Image.Source] = sprite
			batches = append(batches, pixel.NewBatch(&pixel.TrianglesData{}, pictureData))
			batchIndices[tileset.Image.Source] = batchCounter
			batchCounter++
		}
	}
	m.mTiles = batches
	m.mTilesIndices = batchIndices
	m.mTilesCounter = batchCounter
	m.mSprites = sprites
}

func (m *GameMap) CamToTile(x, y int) {
	m.Goto(
		(x*m.mTileWidth)+m.mTileWidth/2,
		(y*m.mTileHeight)+m.mTileHeight/2,
	)
}

func (m *GameMap) Goto(x, y int) {
	m.mCamX = float64(x)
	m.mCamY = float64(y)
}

func (m *GameMap) GetTilePositionAtFeet(x, y int, charW, charH float64) pixel.Vec {
	y = m.mHeight - y
	x = x - 1
	return pixel.V(
		float64(m.mX+(x*m.mTileWidth))-charW,    //x * m.mTileWidth/2
		float64(m.mY+(y*m.mTileHeight))-charH/2, //y * m.mTileHeight/2
	)
}

func getTileLocation(tID, numColumns, numRows int) (x, y int) {
	x = tID % numColumns
	y = numRows - (tID / numColumns) - 1
	return
}

func (m GameMap) getTilePos(idx int) pixel.Vec {
	width := m.mTilemap.Width
	height := m.mTilemap.Height
	gamePos := pixel.V(
		float64(idx%width)-1,
		float64(height)-float64(idx/width),
	)
	return gamePos
}

func (m GameMap) Render() {
	// Draw tiles
	for _, batch := range m.mTiles {
		batch.Clear()
	}

	for _, layer := range m.mTilemap.Layers {
		for tileIndex, tile := range layer.DecodedTiles {
			ts := layer.Tileset
			tID := int(tile.ID)

			if tID == -1 {
				// Tile ID 0 means blank, skip it. temp -1
				continue
			}

			// Calculate the framing for the tile within its tileset's source image
			numRows := ts.Tilecount / ts.Columns
			x, y := getTileLocation(tID, ts.Columns, numRows)
			gamePos := m.getTilePos(tileIndex)

			iX := float64(x) * float64(ts.TileWidth)
			fX := iX + float64(ts.TileWidth)
			iY := float64(y) * float64(ts.TileHeight)
			fY := iY + float64(ts.TileHeight)

			sprite := m.mSprites[ts.Image.Source]
			sprite.Set(sprite.Picture(), pixel.R(iX, iY, fX, fY))
			pos := gamePos.ScaledXY(pixel.V(float64(ts.TileWidth), float64(ts.TileHeight)))
			sprite.Draw(m.mTiles[m.mTilesIndices[ts.Image.Source]], pixel.IM.Moved(pos))
		}
	}

	for _, batch := range m.mTiles {
		batch.Draw(global.gWin)
	}
}

//hero movement
func TeleportCharacter(tileX, tileY int, gMap GameMap, sprite *pixel.Sprite, spriteFrame pixel.Rect) {
	vec := gMap.GetTilePositionAtFeet(tileX, tileY, spriteFrame.W(), spriteFrame.H())
	//set position for sprite
	sprite.Draw(global.gWin, pixel.IM.Moved(vec))
}
