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
	"io/fs"
	"os"
)

//go:embed data/*
var embeddedFiles embed.FS

type GuiData struct {
	BlockSize                      Int  `yaml:"BlockSize"`
	HighlightMoveOk                bool `yaml:"HighlightMoveOk"`
	HighlightMoveNotOk             bool `yaml:"HighlightMoveNotOk"`
	HighlightAttack                bool `yaml:"HighlightAttack"`
	AutoAimAttack                  bool `yaml:"AutoAimAttack"`
	AutoAimAttackFactor            Int  `yaml:"AutoAimAttackFactor"`
	AutoAimMove                    bool `yaml:"AutoAimMove"`
	AutoAimMoveFactor              Int  `yaml:"AutoAimMoveFactor"`
	ShowFreezeCooldownAsMask       bool `yaml:"ShowFreezeCooldownAsMask"`
	ShowMoveCooldownAsMask         bool `yaml:"ShowMoveCooldownAsMask"`
	ShowFreezeCooldownAsBar        bool `yaml:"ShowFreezeCooldownAsBar"`
	ShowMoveCooldownAsBar          bool `yaml:"ShowMoveCooldownAsBar"`
	DrawEnemyHealth                bool `yaml:"DrawEnemyHealth"`
	DrawVirtualCursorDuringReplay  bool `yaml:"DrawVirtualCursorDuringReplay"`
	MoveActualOSCursorDuringReplay bool `yaml:"MoveActualOSCursorDuringReplay"`
	DrawSpawnPortal                bool `yaml:"DrawSpawnPortal"`
	PlaybackBarHeight              Int  `yaml:"PlaybackBarHeight"`
	FrameSkipArrow                 Int  `yaml:"FrameSkipArrow"`
	FrameSkipShiftArrow            Int  `yaml:"FrameSkipShiftArrow"`
	FrameSkipAltArrow              Int  `yaml:"FrameSkipAltArrow"`
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
	FSys                   fs.FS
	buttonRegionWidth      Int
	buttonPause            Rectangle
	buttonNewLevel         Rectangle
	buttonRestartLevel     Rectangle
	buttonPlaybackPlay     Rectangle
	buttonPlaybackBar      Rectangle
	pressedKeys            []ebiten.Key
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
	instructionalText      string
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
	ebiten.SetWindowPosition(1000, 100)

	var g Gui
	g.username = getUsername()
	g.uploadDataChannel = make(chan uploadData)
	go UploadPlaythroughs(g.uploadDataChannel)
	// g.db = ConnectToDbSql()
	// g.world = NewWorld(RInt(I(0), I(10000000)))

	g.textHeight = I(40)
	g.guiMargin = I(50)
	g.buttonRegionWidth = I(200)

	// replayFile := "d:\\Miln\\test.mln999"
	replayFile := ""
	// g.recordingFile = "d:\\Miln\\test.mln999"

	if len(os.Args) == 2 {
		replayFile = os.Args[1]
	}

	if !FileExists(os.DirFS("."), "data") {
		g.FSys = &embeddedFiles
	} else {
		g.FSys = os.DirFS(".")
		g.folderWatcher1.Folder = "data/gui"
		g.folderWatcher2.Folder = "data/levelgenerator"
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
		g.world = NewWorld(g.playthrough.Seed, g.playthrough.Level)
		g.state = Playback
	} else {
		g.playbackExecution = false
		// g.recordingFile = GetNewRecordingFile()
		// seed, targetDifficulty := GetNextLevel(g.username)
		seed := RInt(I(0), I(1000000))
		level := GenerateLevel(g.FSys)
		g.world = NewWorld(seed, level)
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

	return windowSize
}

func (g *Gui) updateWindowSize() {
	// windowSize := g.getWindowSize()
	// ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	ebiten.SetWindowSize(900, 900)
	ebiten.SetWindowTitle("Miln")
}
