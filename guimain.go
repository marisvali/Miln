package main

import (
	"embed"
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	. "github.com/marisvali/miln/gamelib"
	. "github.com/marisvali/miln/world"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	_ "image/png"
	"os"
	"path/filepath"
)

// ReleaseVersion is the version of an executable built and given to someone
// to play, either as a Windows executable or a .wasm on the browser. It is
// meant as a unique label for the functionality that a user/player is presented
// with.
// ReleaseVersion is expected to change very often. Certainly every time a new
// executable is built and sent to someone, it should be tagged with a unique
// ReleaseVersion.
// ReleaseVersion must change when SimulationVersion or InputVersion change.
// But there are many reasons for ReleaseVersion to change when the simulation
// stays the same and the input format stays the same:
// - the executable uses randomly generated levels or fixed levels
// - different fixed levels are included in the executable
// - communication with the server is enabled or disabled
// - graphics change
// All of these changes can be handled by having a generic code that is compiled
// once and depends on a configuration. I very intentionally do not do this.
// The philosophy of this project is currently to release a different executable
// for each variation. The reasons for this:
// - Only the developer is truly comfortable editing the configuration file. If
// the developer has to intervene, he might as well compile a version for the
// user. If it's hard or annoying to compile and release a new version, then
// that process should be improved instead of avoided.
// - I want to keep track of things, who got what experience. The point of a
// configuration is to be able to change things quickly and on the fly. But if
// I want to keep track of things, I must remember to change the release version
// every time I edit the configuration. So the ability to change configurations
// easily is a liability more than a help. It's more helpful, for tracking
// things, to have an unmodifiable binary for each variation.
// - Currently the executables are small enough and I need few enough variations
// that I can easily afford to generate an entire game release for each
// variation (35mb for a Windows .exe and 25mb for a .wasm).
const ReleaseVersion = 999

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

type UserData struct {
	CurrentFixedLevelIdx Int `yaml:"CurrentFixedLevelIdx"`
}

type Gui struct {
	// db                 *sql.DB
	GuiData
	UserData
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
	FSys                   FS
	buttonRegionWidth      Int
	buttonPause            Rectangle
	buttonNewLevel         Rectangle
	buttonNextLevel        Rectangle
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
	visWorld               VisWorld
	layout                 Pt
	playbackPaused         bool
	instructionalText      string
	fixedLevels            []string
	username               string
}

type uploadData struct {
	user              string
	releaseVersion    int64
	simulationVersion int64
	inputVersion      int64
	playthrough       *Playthrough
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
	g.playthrough.InputVersion = I(InputVersion)
	g.playthrough.SimulationVersion = I(SimulationVersion)
	g.playthrough.ReleaseVersion = I(ReleaseVersion)
	g.username = getUsername()
	// A channel size of 10 means the channel will buffer 10 inputs before it is
	// full and it blocks. Hopefully, when uploading data, a size of 10 is
	// sufficient.
	g.uploadDataChannel = make(chan uploadData, 10)
	go UploadPlaythroughs(g.uploadDataChannel)
	// g.db = ConnectToDbSql()
	// g.world = NewWorld(RInt(I(0), I(10000000)))

	g.textHeight = I(40)
	g.guiMargin = I(50)
	g.buttonRegionWidth = I(200)

	// inputFile := "d:\\Miln\\code\\world\\playthroughs\\20250511-091615.mln999-new"
	inputFile := ""
	// g.recordingFile = "d:\\Miln\\test.mln999"

	if len(os.Args) == 2 {
		inputFile = os.Args[1]
	}

	if !FileExists(os.DirFS(".").(FS), "data") {
		g.FSys = &embeddedFiles
	} else {
		g.FSys = os.DirFS(".").(FS)
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

	if FileExists(g.FSys, "data/levels") {
		g.InitializeFixedLevels()
	}

	if inputFile != "" {
		if IsYamlLevel(inputFile) {
			// Play level loaded from YAML file.
			g.playbackExecution = false
			fs := os.DirFS(filepath.Dir(inputFile)).(FS)
			file := filepath.Base(inputFile)
			g.startLevel(LoadLevelFromYAML(fs, file))
		} else {
			// Replay a playthrough loaded from a file.
			g.playbackExecution = true
			g.playthrough = DeserializePlaythrough(ReadFile(inputFile))
			g.world = NewWorldFromPlaythrough(g.playthrough)
			g.state = Playback
		}
	} else if g.UsingFixedLevels() {
		// Play pre-defined levels in order.
		g.playbackExecution = false
		if g.HasMoreFixedLevels() {
			g.startLevel(g.GetCurrentFixedLevel())
		} else {
			// Show a bogus, empty level, just so that the code that draws
			// the interface can work as usual.
			var l Level
			g.world = NewWorld(I(0), l)
			g.state = GameWon
		}
	} else {
		// Play random level.
		g.playbackExecution = false
		g.StartNewLevel()
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
	playSize := IPt(NRows, NCols).Times(g.BlockSize)
	windowSize := playSize
	windowSize.X.Add(g.guiMargin.Times(TWO))
	windowSize.Y.Add(g.guiMargin)
	windowSize.Y.Add(g.textHeight.Times(TWO))

	return windowSize
}

func (g *Gui) updateWindowSize() {
	// windowSize := g.getWindowSize()
	// ebiten.SetWindowSize(windowSize.X.ToInt(), windowSize.Y.ToInt())
	width, height := ebiten.ScreenSizeInFullscreen()
	size := min(width, height) * 8 / 10
	ebiten.SetWindowSize(size, size)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Miln")
}

func (g *Gui) LoadUserData() {
	s := GetUserDataHttp(g.username)
	err := yaml.Unmarshal([]byte(s), &g.UserData)
	Check(err)
}

func (g *Gui) SaveUserData() {
	data, err := yaml.Marshal(g.UserData)
	Check(err)
	SetUserDataHttp(g.username, string(data))
}

func (g *Gui) startLevel(seed Int, l Level) {
	g.playthrough.Seed = seed
	g.playthrough.Level = l
	g.playthrough.Id = uuid.New()
	g.playthrough.History = g.playthrough.History[:0]
	g.world = NewWorldFromPlaythrough(g.playthrough)
	InitializeIdInDbHttp(g.username,
		g.playthrough.ReleaseVersion.ToInt64(),
		g.playthrough.SimulationVersion.ToInt64(),
		g.playthrough.InputVersion.ToInt64(),
		g.playthrough.Id)
	g.state = GameOngoing
}
