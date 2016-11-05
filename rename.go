package main

import (
    "flag"
    "fmt"
    "github.com/dhowden/tag"
    "log"
    "os"
    "path/filepath"
    "strings"
)

func main() {
    var (
        isDryrun bool
        source string
        rootDir string
    )
    flag.BoolVar(&isDryrun, "dryrun", false, "do not move files actually")
    flag.StringVar(&source, "source", "", "specify FLAC/MP3/AAC file or directory containing them")
    flag.StringVar(&rootDir, "root", ".", "specify root directroy into which source files are moved")
    flag.Parse()

    if _,err := os.Stat(source); os.IsNotExist(err) {
        log.Fatal(err)
    }

    if _,err := os.Stat(rootDir); os.IsNotExist(err) {
        log.Fatal(err)
    }

    if err := filepath.Walk(source, RenameFileFunc(rootDir, isDryrun)); err != nil {
        log.Fatal(err)
    }

}

func IsInStrSlice(str string, slice []string) bool {
    for _, elem := range slice {
        if str == elem {
            return true
        }
    }
    return false
}

func RenameFileFunc(rootDir string, isDryrun bool) filepath.WalkFunc {
    supportFileType := []string{".flac", ".mp3", ".m4a", "aac", ".FLAC", ".MP3", ".M4A", "AAC"}

    // Return function of filepath.WalkFunc type.
    // This function is called inside loop iterating files recursively
    return func(path string, info os.FileInfo, err error) error {
        if info.Mode().IsDir() {
            return nil
        }

        if !IsInStrSlice(filepath.Ext(path),supportFileType) {
            return nil
        }

        log.Println(path)

        file,err := os.Open(path)
        if err != nil {
            return err
        }

        audioTag,err := tag.ReadFrom(file)
        if err != nil {
            return err
        }

        relNewPath,err := FilepathFmt(audioTag, filepath.Ext(path))
        if err != nil {
            return err
        }

        newPath := fmt.Sprintf("%s/%s", rootDir, relNewPath)
        log.Printf("===> %s", newPath)

        if _,err := os.Stat(newPath); !os.IsNotExist(err) {
            log.Println("(file already exists. skipping...)")
            return nil
        }

        if isDryrun {
            return nil
        }

        if _,err := os.Stat(filepath.Dir(newPath)); os.IsNotExist(err) {
            os.MkdirAll(filepath.Dir(newPath),0755)
        }

        if err := os.Rename(path, newPath); err != nil {
            log.Println(err)
        }

        return nil
    }
}

func FilepathFmt(tags tag.Metadata, fileExt string) (string, error) {
    discN, discMax := tags.Disc()
    trackN, _ := tags.Track()
    title := tags.Title()
    album := tags.Album()
    artist := tags.Artist()
    albArtist := tags.AlbumArtist()

    var (
        artistDir string
        albumDir string
        discPrfx string
        trackPrfx string
        titleStr string
    )

    albumDir = strings.Replace(album, "/", "_", -1)
    trackPrfx = fmt.Sprintf("%02d - ",trackN)
    titleStr = strings.Replace(title, "/", "_", -1)

    if len(albArtist) > 0  {
        artistDir = strings.Replace(albArtist, "/", "_", -1)
    } else {
        artistDir = strings.Replace(artist, "/", "_", -1)
    }

    if ( discN > 0 || discMax > 0 ) {
        discPrfx = fmt.Sprintf("%d_",discN)
    } else {
        discPrfx = ""
    }

    return fmt.Sprintf("%s/%s/%s%s%s%s", artistDir, albumDir, discPrfx, trackPrfx, titleStr, fileExt), nil

}
