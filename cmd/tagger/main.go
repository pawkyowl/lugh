package main

import (
	"log"
	"path/filepath"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/pawkyowl/lugh/internal/tagger"
)

var (
	scanPath   string = "."
	configFile string = "collection.yaml"
	dry        bool   = false
)

func main() {
	cobra.OnInitialize(func() {
		viper.AutomaticEnv()
	})

	var parseCmd = &cobra.Command{
		Use: "scan",
		Run: parse,
	}

	var applyCmd = &cobra.Command{
		Use: "apply",
		Run: apply,
	}
	applyCmd.PersistentFlags().Bool("dry", dry, "Dry run")
	err := viper.BindPFlag("dry", applyCmd.PersistentFlags().Lookup("dry"))
	if err != nil {
		log.Fatal(err)
	}

	var rootCmd = &cobra.Command{
		Short: "ID3 parser",
	}
	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.PersistentFlags().String("folder", scanPath, "Folder to scan")
	err = viper.BindPFlag("folder", rootCmd.PersistentFlags().Lookup("folder"))
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.PersistentFlags().String("config", configFile, "Config file")
	err = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	if err != nil {
		log.Fatal(err)
	}
	if err = rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func scan(dir string) *tagger.Collection {
	log.Printf("Scanning folder " + dir)
	collection := tagger.Scan(dir)
	log.Printf("%x albums found", len(collection.Albums()))
	return collection
}

func parse(cmd *cobra.Command, args []string) {
	dir := viper.GetString("folder")
	collection := scan(dir)
	if len(collection.Albums()) > 0 {
		for _, key := range collection.Albums() {
			album := collection.GetAlbum(key)
			err := album.SavePicture()
			if err != nil {
				log.Fatal(err)
			}
		}
		err := collection.Save(filepath.Join(dir, viper.GetString("config")))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func apply(cmd *cobra.Command, args []string) {
	dir := viper.GetString("folder")
	newCollection, err := tagger.LoadCollection(filepath.Join(dir, viper.GetString("config")))
	if err != nil {
		log.Fatal(err)
	}
	collection := scan(dir)
	if viper.GetBool("dry") {
		for _, key := range collection.Albums() {
			album := collection.GetAlbum(key)
			newAlbum := newCollection.GetAlbum(key)
			if newAlbum == nil {
				log.Fatal(key + " mismatch")
			}
			if album.Path != newAlbum.Path {
				log.Fatal(key + " mismatch")
			}
			album.Compare(newAlbum)
		}
	} else {
		for _, key := range collection.Albums() {
			album := collection.GetAlbum(key)
			newAlbum := newCollection.GetAlbum(key)
			if newAlbum == nil {
				log.Fatal(key + " mismatch")
			}
			if album.Path != newAlbum.Path {
				log.Fatal(key + " mismatch")
			}
			album.Copy(newAlbum)
			for _, track := range album.Tracks {
				err = track.Save()
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
