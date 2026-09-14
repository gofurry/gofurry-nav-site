// asset-cutover is an offline maintenance tool, never invoked by serve or Goose.
// It stages immutable objects and emits SQL; it never connects to a database.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	env "github.com/gofurry/gofurry-admin/config"
	"github.com/gofurry/gofurry-admin/internal/infra/assets"
)

type siteIcon struct {
	SiteID int64   `json:"site_id"`
	Old    *string `json:"old"`
	New    string  `json:"new"`
}
type hero struct {
	Name      string `json:"name"`
	ObjectKey string `json:"object_key"`
}
type manifest struct {
	SiteIcons   []siteIcon `json:"site_icons"`
	HeroDesktop []hero     `json:"hero_desktop"`
	HeroMobile  []hero     `json:"hero_mobile"`
}
type sources struct {
	SiteIcons []struct {
		SiteID int64   `json:"site_id"`
		Old    *string `json:"old"`
		File   string  `json:"file"`
	} `json:"site_icons"`
	HeroDesktop []sourceHero `json:"hero_desktop"`
	HeroMobile  []sourceHero `json:"hero_mobile"`
}
type sourceHero struct {
	Name string `json:"name"`
	File string `json:"file"`
}

func readJSON(path string, value any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	d := json.NewDecoder(io.LimitReader(f, 16<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected exactly one JSON document")
	}
	return nil
}

func prepare(sourceFile string) (manifest, []assets.Object, error) {
	m := manifest{SiteIcons: []siteIcon{}, HeroDesktop: []hero{}, HeroMobile: []hero{}}
	var source sources
	if err := readJSON(sourceFile, &source); err != nil {
		return m, nil, err
	}
	objects := []assets.Object{}
	load := func(kind string, id int64, file string) (assets.Object, error) {
		path := file
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(sourceFile), path)
		}
		f, err := os.Open(path)
		if err != nil {
			return assets.Object{}, err
		}
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, assets.MaxSize+1))
		if err != nil {
			return assets.Object{}, err
		}
		return assets.NewObject(kind, id, filepath.Base(path), data)
	}
	for _, item := range source.SiteIcons {
		o, err := load("site-icon", item.SiteID, item.File)
		if err != nil {
			return m, nil, err
		}
		objects = append(objects, o)
		m.SiteIcons = append(m.SiteIcons, siteIcon{item.SiteID, item.Old, o.Key})
	}
	for _, pool := range []struct {
		variant string
		items   []sourceHero
		out     *[]hero
	}{{"desktop", source.HeroDesktop, &m.HeroDesktop}, {"mobile", source.HeroMobile, &m.HeroMobile}} {
		for _, item := range pool.items {
			o, err := load("hero-"+pool.variant, 0, item.File)
			if err != nil {
				return m, nil, err
			}
			objects = append(objects, o)
			*pool.out = append(*pool.out, hero{item.Name, o.Key})
		}
	}
	return m, objects, validateManifest(m)
}

func run() error {
	var config, source, input, output string
	flag.StringVar(&config, "config", "", "explicit Admin YAML (required for staging; no .env loading)")
	flag.StringVar(&source, "sources", "", "source file list; upload only, never modify database references")
	flag.StringVar(&input, "manifest", "", "already staged manifest; generate SQL without cloud access")
	flag.StringVar(&output, "output-dir", "", "new directory for manifest, publication report and SQL")
	flag.Parse()
	if flag.NArg() != 0 || output == "" || (source == "") == (input == "") {
		return errors.New("specify --output-dir and exactly one of --sources or --manifest")
	}
	var m manifest
	var objects []assets.Object
	var err error
	if source != "" {
		if config == "" {
			return errors.New("--config is required with --sources")
		}
		m, objects, err = prepare(source)
	} else {
		err = readJSON(input, &m)
	}
	if err != nil {
		return err
	}
	cutover, rollback, err := renderSQL(m)
	if err != nil {
		return err
	}
	if err := os.Mkdir(output, 0700); err != nil {
		return fmt.Errorf("output directory must be new: %w", err)
	}
	publications := []assets.Publication{}
	if source != "" {
		if err := env.MustInitServerConfig("gofurry-admin", config); err != nil {
			return errors.New("cannot load explicit Admin configuration")
		}
		storage, err := assets.New(env.GetServerConfig().ExternalServices.AssetStorage)
		if err != nil {
			return err
		}
		for _, object := range objects {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			result, err := storage.Publish(ctx, object)
			cancel()
			if err != nil {
				return fmt.Errorf("stage %s: %w", object.Key, err)
			}
			publications = append(publications, result)
			fmt.Printf("%s: COS=%s R2=%s\n", object.Key, result.Primary, result.Mirror)
		}
	}
	data, _ := json.MarshalIndent(m, "", "  ")
	report, _ := json.MarshalIndent(publications, "", "  ")
	for name, body := range map[string][]byte{"manifest.json": append(data, '\n'), "publication.json": append(report, '\n'), "cutover.sql": []byte(cutover), "rollback.sql": []byte(rollback)} {
		if err := os.WriteFile(filepath.Join(output, name), body, 0600); err != nil {
			return err
		}
	}
	fmt.Println("Prepared files. Review SQL and publication warnings before the maintenance window; no database changes were executed.")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
