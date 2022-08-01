package bookstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	_ "github.com/go-sql-driver/mysql"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBookstack(t *testing.T) {

	check := require.New(t)

	ctx := context.Background()

	host := "bookstack"
	hostDB := fmt.Sprintf("%s_db", host)
	hostNet := fmt.Sprintf("%s_net", host)
	waitFor := wait.ForLog("s6-rc: info: service 99-ci-service-check successfully started")

	db := struct {
		user     string
		password string
		database string
	}{
		user:     fake.UserName(),
		password: fake.SimplePassword(),
		database: fake.Word(),
	}

	n, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: hostNet,
		},
	})

	check.NoError(err)
	defer n.Remove(ctx)

	// MariaDB Container
	mariaReq := testcontainers.ContainerRequest{
		Image:        "lscr.io/linuxserver/mariadb:latest",
		AutoRemove:   true,
		ExposedPorts: []string{"3306/tcp"},
		Name:         hostDB,
		Env: map[string]string{
			"MYSQL_DATABASE":      db.database,
			"MYSQL_PASSWORD":      db.password,
			"MYSQL_ROOT_PASSWORD": db.password,
			"MYSQL_USER":          db.user,
			"PGID":                "1000",
			"PUID":                "1000",
			"TZ":                  "Europe/London",
		},
		Networks: []string{hostNet},
		WaitingFor: wait.ForAll(
			waitFor,
			wait.ForListeningPort("3306/tcp"),
		),
	}

	maria, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: mariaReq,
		Started:          true,
		Logger:           testcontainers.TestLogger(t),
	})

	check.NoError(err)
	defer maria.Terminate(ctx)

	mariaLogs, err := maria.Logs(ctx)
	check.NoError(err)

	rl, err := ioutil.ReadAll(mariaLogs)
	check.NoError(err)
	check.NotContains(string(rl), "error")

	// Bookstack Container
	bookstackReq := testcontainers.ContainerRequest{
		Image:      "lscr.io/linuxserver/bookstack:v22.07-ls28",
		AutoRemove: true,
		Name:       "bookstack",
		Env: map[string]string{
			"PUID":        "1000",
			"PGID":        "1000",
			"APP_URL":     fmt.Sprintf("http://%s", host),
			"DB_HOST":     hostDB,
			"DB_USER":     db.user,
			"DB_PASS":     db.password,
			"DB_DATABASE": db.database,
		},
		Networks:     []string{hostNet},
		ExposedPorts: []string{"80/tcp"},
		WaitingFor: wait.ForAll(
			waitFor,
			wait.ForListeningPort("80/tcp"),
		),
	}

	stack, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: bookstackReq,
		Logger:           testcontainers.TestLogger(t),
		Started:          true,
	})

	check.NoError(err)

	defer stack.Terminate(ctx)

	booklogs, err := stack.Logs(ctx)
	check.NoError(err)

	booklogString, err := ioutil.ReadAll(booklogs)
	check.NoError(err)

	bls := string(booklogString)

	check.NotContains(bls, "Name does not resolve")
	check.NotContains(bls, "SQLSTATE[HY000] [2002]")

	outside, err := stack.PortEndpoint(ctx, "80", "http")
	check.NoError(err)

	// Browser
	chromiumReq := testcontainers.ContainerRequest{
		Image:        "montferret/chromium:latest",
		AutoRemove:   true,
		Networks:     []string{hostNet},
		ExposedPorts: []string{"9222/tcp"},
		WaitingFor:   wait.ForListeningPort("9222/tcp"),
	}

	chromium, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: chromiumReq,
		Logger:           testcontainers.TestLogger(t),
		Started:          true,
	})

	check.NoError(err)

	defer chromium.Terminate(ctx)

	chromiumAddress, err := chromium.PortEndpoint(ctx, "9222/tcp", "http")
	check.NoError(err)

	u, err := launcher.ResolveURL(chromiumAddress)
	check.NoError(err)
	check.NotEmpty(u)

	browser := rod.New().ControlURL(u)

	check.NoError(browser.Connect())

	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: fmt.Sprintf("http://%s/login", host)})
	check.NoError(err)

	defer page.Close()

	check.NoError(page.WaitLoad())

	check.NoError(page.MustElement("#email").Input("admin@admin.com"))
	check.NoError(page.MustElement("#password").Input("password"))
	check.NoError(page.MustElement(".button").Click(proto.InputMouseButtonLeft))

	check.NoError(page.WaitLoad())

	usr, err := page.MustElement(".dropdown-menu > li:nth-child(3) > a:nth-child(1)").Attribute("href")
	check.NoError(err)

	check.NoError(page.Navigate(path.Join(*usr, "create-api-token")))

	check.NoError(page.WaitLoad())

	check.NoError(page.MustElement("#name").Input("go-bookstack_test"))
	check.NoError(page.MustElement("button.button").Click(proto.InputMouseButtonLeft))

	check.NoError(page.WaitLoad())

	// GET ID
	id, err := page.MustElement("#token_id").Text()
	check.NoError(err)
	check.NotEmpty(id)

	// GET SECRET
	secret, err := page.MustElement("div.grid:nth-child(2) > div:nth-child(2) > input:nth-child(1)").Text()
	check.NoError(err)
	check.NotEmpty(secret)

	t.Run("testing user commands", func(t *testing.T) {

		bk := New(
			SetToken(id, secret),
			SetURL(outside),
			SetLogger(log.Default()),
		)

		users, err := bk.ListUsers(ctx, nil)
		check.NoError(err)
		check.Len(users, 2)

		userParams := UserParams{
			Name:     fake.FullName(),
			Email:    fake.EmailAddress(),
			Password: fake.SimplePassword(),
			Language: "en",
			Roles:    []int{3},
		}

		created, err := bk.CreateUser(ctx, userParams)
		check.NoError(err)
		check.NotEmpty(created)

		check.Equal(userParams.Email, created.Email)

		usr, err := bk.GetUser(ctx, created.ID)
		check.NoError(err)
		check.Equal(created.Email, usr.Email)

		updated, err := bk.UpdateUser(ctx, created.ID, UserParams{Email: fake.EmailAddress()})
		check.NoError(err)
		check.NotEmpty(updated)
		check.NotEqual(updated, created)
		check.NotEqual(updated.Email, created.Email)

		users, err = bk.ListUsers(ctx, nil)
		check.NoError(err)
		check.Len(users, 3)

		ok, err := bk.DeleteUser(ctx, created.ID, nil)
		check.NoError(err)
		check.True(ok)

		users, err = bk.ListUsers(ctx, nil)
		check.NoError(err)
		check.Len(users, 2)

		users, err = bk.ListUsers(ctx, &QueryParams{Count: 1})
		check.NoError(err)
		check.Len(users, 1)

	})

	t.Run("testing book commands", func(t *testing.T) {

		bk := New(
			SetToken(id, secret),
			SetURL(outside),
			SetLogger(log.Default()),
		)

		books, err := bk.ListBooks(ctx, nil)
		check.NoError(err)
		check.Len(books, 0)

		bookParams := BookParams{
			Name:  fake.Title(),
			Image: "./test_data/upload.png",
		}

		created, err := bk.CreateBook(ctx, bookParams)
		check.NoError(err)
		check.NotEmpty(created)
		check.Equal(bookParams.Name, created.Name)

		book, err := bk.GetBook(ctx, created.ID)
		check.NoError(err)
		check.Equal(book.Name, created.Name)
		check.NotEmpty(book.Cover.URL)

		updated, err := bk.UpdateBook(ctx, created.ID, BookParams{Name: fake.Brand()})
		check.NoError(err)
		check.NotEmpty(updated)
		check.NotEqual(created.Name, updated.Name)
		check.NotEqual(created, updated)

		books, err = bk.ListBooks(ctx, nil)
		check.NoError(err)
		check.Len(books, 1)

		ok, err := bk.DeleteBook(ctx, created.ID)
		check.NoError(err)
		check.True(ok)

	})

	// DELETE TOKEN
	check.NoError(page.Navigate(page.MustInfo().URL + "/delete"))
	check.NoError(page.MustElement("#main-content > div > div > div > div > form > button").Click(proto.InputMouseButtonLeft))

}
