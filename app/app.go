package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

type Song struct {
	TrackName  string `json:"track_name"`
	ImageLink  string `json:"image_link"`
	ArtistName string `json:"artist_name"`
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Predict(songLyrics string) string {
	pythonScriptPath := "./call_model.py"
	pythonModelPath := "./model.pkl"
	Pythoncmd := exec.Command(pythonScriptPath, "--song-lyrics", songLyrics, "--model-name", pythonModelPath)

	pythonout, err := Pythoncmd.Output()
	if err != nil {
		fmt.Print(err.Error())
		panic(err)
	}

	var label string
	output := strings.TrimSpace(string(pythonout))

	// This is needed since spotifyapi prefers hip-hop over rap
	if output == "rap" {
		label = "hip-hop"
	} else {
		label = output
	}

	return label
}

func (a *App) GetSpotifyRecommendations(songGenre string) string {
	err := godotenv.Load()

	if err != nil {
		fmt.Print(err.Error())
		panic(err)
	}

	ClientID := os.Getenv("SPOTIFY_ID")
	ClientSecret := os.Getenv("SPOTIFY_SECRET")

	fmt.Println(ClientID)
	fmt.Println(ClientSecret)


	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	token, err := config.Token(ctx)
	if err != nil {
		fmt.Print(err.Error())
		panic(err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)

	fmt.Print(songGenre)
	genres := []string{songGenre}

	seeds := spotify.Seeds{
		Genres: genres,
	}

	track_attributes := spotify.NewTrackAttributes().
		MaxValence(0.4).
		TargetEnergy(0.6).
		TargetDanceability(0.6)

	res, err := client.GetRecommendations(ctx, seeds, track_attributes, spotify.Country("US"), spotify.Limit(3))

	recommendedTracks := []Song{}

	fmt.Print(res.Tracks)
	for _, track := range res.Tracks {
		// fmt.Println(track.Name + " by " + track.Artists[0].Name)

		newSong := Song{
			TrackName:  track.Name,
			ImageLink:  track.Album.Images[0].URL,
			ArtistName: track.Artists[0].Name,
		}
		recommendedTracks = append(recommendedTracks, newSong)

	}

	jsonData, err := json.Marshal(recommendedTracks)
	if err != nil {
		fmt.Print(err.Error())
		panic(err)

	}
	return string(jsonData)
}