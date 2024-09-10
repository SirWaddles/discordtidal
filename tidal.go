package discordtidal

import (
	"fmt"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/unickorn/discordtidal/log"
	"github.com/unickorn/discordtidal/rpc"
	"github.com/unickorn/discordtidal/song"
	"github.com/unickorn/discordtidal/tidal"
)

type Status uint8

const (
	Closed Status = iota
	Opened
	Playing
	Paused
)

var (
	sleepTime       = time.Second * 5
	coverUpdateTime = 0
	t               *tidal.Track
)

// Start starts the Discord RPC update loop.
func Start() {
	defer log.Log().Sync()
	rpc.Init()
	defer rpc.Logout()

	for {
		getSong()

		// if closed, log out
		if status == Closed {
			log.Log().Debugln("Status: CLOSED")
			rpc.Logout()
			sleepTime = time.Second * 5
			time.Sleep(sleepTime)
			continue
		}

		if status == Playing {
			log.Log().Debugln("Status: PLAYING")
			// NEW SONG
			songChanged := song.Current != nil && (song.Current.Track.Title != track || !song.Current.Track.ArtistMatches(artist))
			// song exists, current unix timestamp is bigger than song start time + paused time + duration + 1
			looped := song.Current != nil && time.Now().Unix() > int64(song.Current.Track.Duration)+song.Current.StartTime+int64(song.Current.PausedTime)

			// newly started playing || song changed || looped
			if song.Current == nil || songChanged || looped {
				rpc.Login()
				log.Log().Debugln("[TRIGGER] NEW SONG/CHANGE/LOOP")
				now := time.Now()
				if looped {
					t = song.Current.Track
				} else {
					t = tidal.GetTrack(track, artist)
				}
				song.Current = &song.Song{
					StartTime: now.Unix(),
					Track:     t,
				}

				sleepTime = time.Second

				trackUrl := fmt.Sprintf("https://listen.tidal.com/album/%d/track/%d", song.Current.Track.Album.ID , song.Current.Track.Id)

				log.Log().Infoln("[TRIGGER] BUTTON URL CHANGED TO:", trackUrl)


				// set activity
				end := time.Unix(int64(song.Current.Track.Duration)+song.Current.StartTime+int64(song.Current.PausedTime), 0)
				err := client.SetActivity(client.Activity{
					Details:    song.Current.Track.Title,
					State:      "by " + song.Current.Track.FormatArtists(),
					LargeImage: "tidal",
					LargeText:  song.Current.Track.Album.Title,
					Timestamps: &client.Timestamps{
						Start: &now,
						End:   &end,
					},
					Buttons: []*client.Button{
						{
							Label: "Listen on TIDAL",
							Url:   trackUrl,
						},
					},
				})
				if err != nil {
					panic(err)
				}
			}

			// used to be paused || needs cover update
			if song.Current.Paused || coverUpdateTime == -1 {
				// not paused obviously
				song.Current.Paused = false

				log.Log().Debugln("[TRIGGER] COVER UPDATE/UNPAUSE")

				trackUrl := fmt.Sprintf("https://listen.tidal.com/album/%d/track/%d", song.Current.Track.Album.ID , song.Current.Track.Id)

				start := time.Unix(song.Current.StartTime, 0)
				end := time.Unix(int64(uint64(song.Current.Track.Duration)+uint64(song.Current.StartTime)+song.Current.PausedTime), 0)
				err := client.SetActivity(client.Activity{
					Details:    song.Current.Track.Title,
					State:      "by " + song.Current.Track.FormatArtists(),
					LargeImage: "tidal",
					LargeText:  song.Current.Track.Album.Title,
					Timestamps: &client.Timestamps{
						Start: &start,
						End:   &end,
					},
					Buttons: []*client.Button{
						{
							Label: "Listen on TIDAL",
							Url:   trackUrl,
						},
					},
				})
				if err != nil {
					panic(err)
				}
			}
		}

		// Paused a song
		if status == Paused && song.Current != nil {
			rpc.Login()
			log.Log().Debugln("Status: PAUSED")
			song.Current.PausedTime += uint64(sleepTime / time.Second)

			trackUrl := fmt.Sprintf("https://listen.tidal.com/album/%d/track/%d", song.Current.Track.Album.ID , song.Current.Track.Id)

			sleepTime = time.Second * 2
			if !song.Current.Paused {
				song.Current.Paused = true
				err := client.SetActivity(client.Activity{
					Details:    song.Current.Track.Title,
					State:      "by " + song.Current.Track.FormatArtists(),
					LargeImage: "tidal",
					LargeText:  song.Current.Track.Album.Title,
					Buttons: []*client.Button{
						{
							Label: "Listen on TIDAL",
							Url:   trackUrl,
						},
					},
				})
				if err != nil {
					panic(err)
				}
			}
		}

		time.Sleep(sleepTime)
	}
}
