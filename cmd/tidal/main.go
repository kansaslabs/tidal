package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rotationalio/tidal"
	"gopkg.in/urfave/cli.v1"
)

const (
	usageText = `tidal [-p PACKAGE] [-m DIR]

   The default tidal command generates migrations; common usage is to add
   a go generate directive in a package and use the go generate command to
   build migrations directly into your source code.

   Note that migrations are discovered either by looking for a "migrations"
   directory in the current working directory or using a specified directory
   as an argument. The utility falls back to the current working directory.

   Tidal also has several utility and helper commands:

   tidal command [command options] [args ...]`

	newUsageText = `tidal new [-n "name of migration"] [-p PACKAGE] [-m DIR]

   Creates a new migration file in the specified directory, otherwise looks
   for a "migrations" directory, then defaults to the current working directory.`

	migrateUsageText = `tidal migrate [-D] [-m DIR] [-r REVISION] [-d URL]

   A helper utility to test migration SQL before embedding them.
   This command checks the current migration status in the database and
   applies all migrations in the specified directory (or "migrations" or
   CWD) up to the specified or latest revision.`

	rollbackUsageText = `tidal rollback [-D] [-m DIR] [-r REVISION] [-d URL]

   A helper utility to test migration SQL before embedding them.
   This command checks the current migration status in the database and
   rolls back all migrations in the specified directory (or "migrations" or
   CWD) down to the specified or all the way back to no-migrations.`
)

func main() {
	app := cli.NewApp()
	app.Name = "tidal"
	app.Version = tidal.Version()
	app.Usage = "generate and manage migrations from SQL files"
	app.UsageText = usageText
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "p, package",
			Usage: "override package name from directory or migration file(s)",
		},
		cli.StringFlag{
			Name:  "m, migrations",
			Usage: "specify directory to look for migrations in (otherwise performs search)",
		},
		cli.StringFlag{
			Name:  "o, out",
			Usage: "location to write generated code (default: migrations parent directory)",
		},
	}
	app.Action = generate
	app.Commands = []cli.Command{
		{
			Name:      "new",
			Usage:     "create a new blank migration file",
			UsageText: newUsageText,
			Action:    create,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "n, name",
					Usage: "name or title or the revision (default auto_datetime)",
				},
				cli.StringFlag{
					Name:  "p, package",
					Usage: "specify the name of the package in the template",
				},
				cli.StringFlag{
					Name:  "m, migrations",
					Usage: "specify directory to create migration in (otherwise performs search)",
				},
			},
		},
		{
			Name:   "revision",
			Usage:  "display the current migration status of the database",
			Action: revision,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the database uri to connect to",
					EnvVar: "DATABASE_URL",
				},
				cli.IntFlag{
					Name:  "r, revision",
					Usage: "specify a revision to get the detail status for",
					Value: -1,
				},
			},
		},
		{
			Name:      "migrate",
			Aliases:   []string{"up"},
			Usage:     "apply migrations from the specified migrations directory",
			UsageText: migrateUsageText,
			Action:    migrate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "m, migrations",
					Usage: "specify directory to look for migrations in (otherwise performs search)",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the database uri to connect to",
					EnvVar: "DATABASE_URL",
				},
				cli.IntFlag{
					Name:  "r, revision",
					Usage: "specify a revision to migrate up to (otherwise applies all)",
					Value: -1,
				},
				cli.BoolFlag{
					Name:  "D, debug",
					Usage: "specify migration actions without actually executing them",
				},
			},
		},
		{
			Name:      "rollback",
			Aliases:   []string{"down"},
			Usage:     "rollback migrations from the specified migrations directory",
			UsageText: rollbackUsageText,
			Action:    rollback,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "m, migrations",
					Usage: "specify directory to look for migrations in (otherwise performs search)",
				},
				cli.StringFlag{
					Name:   "d, db",
					Usage:  "the database uri to connect to",
					EnvVar: "DATABASE_URL",
				},
				cli.IntFlag{
					Name:  "r, revision",
					Usage: "specify a revision to rollback down to (otherwise rollsback all)",
					Value: -1,
				},
				cli.BoolFlag{
					Name:  "D, debug",
					Usage: "specify rollback actions without actually executing them",
				},
			},
		},
	}

	// Run the program, it should not error
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func generate(c *cli.Context) (err error) {
	var mdir string
	if mdir, err = findMigrations(c); err != nil {
		return cli.NewExitError(err, 1)
	}

	outpath := determineFileOutputPath(c)
	packageName := c.String("package")

	if err = tidal.Generate(mdir, outpath, packageName); err != nil {
		return cli.NewExitError(err, 1)
	}
	return nil
}

func create(c *cli.Context) (err error) {
	return nil
}

func revision(c *cli.Context) (err error) {
	fmt.Println("rvision")
	return nil
}

func migrate(c *cli.Context) (err error) {
	fmt.Println("migrate")
	return nil
}

func rollback(c *cli.Context) (err error) {
	fmt.Println("migrate")
	return nil
}

// helper utility to search for migrations directory
func findMigrations(c *cli.Context) (path string, err error) {
	if path = c.String("migrations"); path != "" {
		return path, nil
	}

	var cwd string
	if cwd, err = os.Getwd(); err != nil {
		return "", err
	}

	dirs := make([]string, 0)
	if err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			basename := strings.ToLower(info.Name())
			if basename == "migrations" {
				dirs = append(dirs, path)
				return nil
			}

			if strings.HasPrefix(basename, ".") || strings.HasPrefix(basename, "~") {
				return filepath.SkipDir
			}
		}
		return nil
	}); err != nil {
		return "", err
	}

	switch len(dirs) {
	case 0:
		return filepath.Rel(cwd, cwd)
	case 1:
		return filepath.Rel(cwd, dirs[0])
	default:
		return "", fmt.Errorf("discovered %d migrations directories, please specify which one to use", len(dirs))
	}
}

// If outpath is a go file, e.g. ends in .go - simply write it to that file. Otherwise,
// assume it is a directory. If the basename is "migrations" use the parent directory.
func determineFileOutputPath(c *cli.Context) (outpath string) {
	outpath = c.String("out")

	if strings.HasSuffix(outpath, ".go") {
		return outpath
	}

	if strings.ToLower(filepath.Base(outpath)) == "migrations" {
		return filepath.Join(filepath.Dir(outpath), "migrations.go")
	}

	return filepath.Join(outpath, "migrations.go")
}
