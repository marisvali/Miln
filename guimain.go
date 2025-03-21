package main

import (
	"embed"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	. "github.com/marisvali/miln/ai"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	_ "image/png"
	"os"
)

//go:embed data/*
var embeddedFiles embed.FS

type GuiData struct {
	BlockSize                      Int
	HighlightMoveOk                bool
	HighlightMoveNotOk             bool
	HighlightAttack                bool
	AutoAimAttack                  bool
	AutoAimAttackFactor            Int
	AutoAimMove                    bool
	AutoAimMoveFactor              Int
	ShowFreezeCooldownAsMask       bool
	ShowMoveCooldownAsMask         bool
	ShowFreezeCooldownAsBar        bool
	ShowMoveCooldownAsBar          bool
	DrawEnemyHealth                bool
	DrawVirtualCursorDuringReplay  bool
	MoveActualOSCursorDuringReplay bool
	DrawSpawnPortal                bool
	PlaybackBarHeight              Int
}

type Animations struct {
	animMoveFailed             Animation
	animAttackFailed           Animation
	animPlayer                 Animation
	animPlayer1                Animation
	animPlayer2                Animation
	animHoundSearching         Animation
	animHoundPreparingToAttack Animation
	animHoundAttacking         Animation
	animHoundHit               Animation
	animHoundDead              Animation
}

type Gui struct {
	// db                 *sql.DB
	GuiData
	Animations
	defaultFont           font.Face
	imgGround             *ebiten.Image
	imgTree               *ebiten.Image
	imgPlayerHealth       *ebiten.Image
	imgPlayerAmmo         *ebiten.Image
	imgHound              *ebiten.Image
	imgEnemyHealth        *ebiten.Image
	imgEnemyCooldown      *ebiten.Image
	imgTileOverlay        *ebiten.Image
	imgBeam               *ebiten.Image
	imgShadow             *ebiten.Image
	imgTextBackground     *ebiten.Image
	imgTextColor          *ebiten.Image
	imgAmmo               *ebiten.Image
	imgSpawnPortal        *ebiten.Image
	imgPlayerHitEffect    *ebiten.Image
	imgHighlightMoveOk    *ebiten.Image
	imgHighlightMoveNotOk *ebiten.Image
	imgHighlightAttack    *ebiten.Image
	imgBlack              *ebiten.Image
	imgCursor             *ebiten.Image
	imgPlayBar            *ebiten.Image
	imgPlaybackPlay       *ebiten.Image
	imgPlaybackPause      *ebiten.Image
	imgPlaybackCursor     *ebiten.Image

	world                  World
	frameIdx               Int
	folderWatcher1         FolderWatcher
	folderWatcher2         FolderWatcher
	playbackExecution      bool
	recordingFile          string
	state                  GameState
	textHeight             Int
	guiMargin              Int
	EmbeddedFS             *embed.FS
	buttonRegionWidth      Int
	buttonPause            Rectangle
	buttonNewLevel         Rectangle
	buttonRestartLevel     Rectangle
	buttonPlaybackPlay     Rectangle
	buttonPlaybackBar      Rectangle
	justPressedKeys        []ebiten.Key // keys pressed in this frame
	mousePt                Pt           // mouse position in this frame
	leftButtonJustPressed  bool         // left mouse button state in this frame
	rightButtonJustPressed bool         // right mouse button state in this frame
	playerHitEffectIdx     Int
	playthrough            Playthrough
	uploadDataChannel      chan uploadData
	username               string
	ai                     AI
	visWorld               VisWorld
	layout                 Pt
	playbackPaused         bool
}

type uploadData struct {
	user        string
	version     int64
	id          uuid.UUID
	playthrough []byte
}

type GameState int64

const (
	GameOngoing GameState = iota
	GamePaused
	GameWon
	GameLost
	Playback
)

func main() {
	ebiten.SetWindowPosition(10, 100)

	var g Gui
	g.username = getUsername()
	g.uploadDataChannel = make(chan uploadData)
	go UploadPlaythroughs(g.uploadDataChannel)
	// g.db = ConnectToDbSql()
	// g.world = NewWorld(RInt(I(0), I(10000000)))

	g.textHeight = I(40)
	g.guiMargin = I(50)
	g.buttonRegionWidth = I(200)

	replayFile := "recordings/recorded-inputs-2025-03-21-000000.mln"
	// replayFile := ""
	// replayFile := "d:\\gms\\Miln\\analysis\\tools\\denis\\20250311-213937.mln008"

	if len(os.Args) == 2 {
		replayFile = os.Args[1]
	}

	if !FileExists("data") {
		g.EmbeddedFS = &embeddedFiles
	} else {
		g.folderWatcher1.Folder = "data/gui"
		g.folderWatcher2.Folder = "data/world"
		// Initialize watchers.
		// Check if folder contents changed but do nothing with the result
		// because we just want the watchers to initialize their internal
		// structures with the current timestamps of files.
		// This is necessary if we want to avoid creating a new world
		// immediately after the first world is created, every time.
		// I want to avoid creating a new world for now because it changes the
		// id of the current world and it messes up the upload of the world
		// to the database.
		g.folderWatcher1.FolderContentsChanged()
		g.folderWatcher2.FolderContentsChanged()
	}

	if replayFile != "" {
		g.playbackExecution = true
		g.playthrough = DeserializePlaythrough(ReadFile(replayFile))
		g.world = NewWorld(g.playthrough.Seed, g.playthrough.TargetDifficulty, g.EmbeddedFS)
		g.state = Playback
	} else {
		g.playbackExecution = false
		g.recordingFile = GetNewRecordingFile()
		// seed, targetDifficulty := GetNextLevel(g.username)
		seed, targetDifficulty := RInt(I(0), I(1000000)), RInt(I(60), I(70))
		g.world = NewWorld(seed, targetDifficulty, g.EmbeddedFS)
		// g.world = NewWorld(RInt(I(0), I(1000000)))
		// InitializeIdInDbSql(g.db, g.world.Id)
		// UploadDataToDbSql(g.db, g.world.Id, g.world.SerializedPlaythrough())
		InitializeIdInDbHttp(g.username, Version, g.world.Id)
		g.state = GameOngoing
	}

	g.loadGuiData()
	g.imgTileOverlay = ebiten.NewImage(g.BlockSize.ToInt(), g.BlockSize.ToInt())

	// Load the Arial font.
	var err error
	fontData, err := opentype.Parse(goregular.TTF)
	Check(err)

	g.defaultFont, err = opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    24,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	Check(err)

	// Start the game.
	err = ebiten.RunGame(&g)
	Check(err)
}

func (g *Gui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.layout = g.getWindowSize()
	return g.layout.X.ToInt(), g.layout.Y.ToInt()
}

func (g *Gui) getWindowSize() Pt {
	playSize := g.world.Obstacles.Size().Times(g.BlockSize)
	windowSize := playSize
	windowSize.X.Add(g.guiMargin.Times(TWO))
	windowSize.Y.Add(g.guiMargin)
	windowSize.Y.Add(g.textHeight.Times(TWO))
	if g.playbackExecution {
		windowSize.Y.Add(g.textHeight)
	}

	return windowSize
}

func (g *Gui) updateWindowSize() {
	// windowSize := g.getWindowSize()
	// ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	ebiten.SetWindowSize(900, 900)
	ebiten.SetWindowTitle("Miln")
}
