package bookstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"path"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/icrowley/fake"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestBookstack(t *testing.T) {

	check := require.New(t)

	ctx := context.Background()

	img := "./test_data/upload.png"

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
			Name:     fmt.Sprintf("%s %s", fake.FirstName(), fake.LastName()),
			Email:    fake.EmailAddress(),
			Password: fake.Password(10, 24, true, true, true),
			Language: "en",
			Roles:    []int{3},
		}

		created, err := bk.CreateUser(ctx, userParams)
		check.NoError(err, "unable to create user: %+v\n", userParams)
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
			Image: img,
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

		plain, err := bk.ExportBookPlaintext(ctx, created.ID)
		check.NoError(err)
		check.NotEmpty(plain)

		b, err := ioutil.ReadAll(plain)
		check.NoError(err)
		check.Contains(string(b), updated.Name)

		_, err = bk.ExportBookHTML(ctx, created.ID)
		check.NoError(err)

		_, err = bk.ExportBookPDF(ctx, created.ID)
		check.NoError(err)

		_, err = bk.ExportBookMarkdown(ctx, created.ID)
		check.NoError(err)

		ok, err := bk.DeleteBook(ctx, created.ID)
		check.NoError(err)
		check.True(ok)

	})

	t.Run("entire process", func(t *testing.T) {

		bk := New(
			SetToken(id, secret),
			SetURL(outside),
			SetLogger(log.Default()),
		)

		ctx := context.Background()

		up := UserParams{
			Name:     fmt.Sprintf("%s %s", fake.FirstName(), fake.LastName()),
			Email:    fake.EmailAddress(),
			Password: fake.Password(10, 24, true, true, true),
		}

		user, err := bk.CreateUser(ctx, up)
		check.NoError(err, "unable to create user: %+v\n", up)
		check.NoError(err)

		usr, err := bk.GetUser(ctx, user.ID)
		check.NoError(err)
		check.Equal(user, usr)

		usr, err = bk.UpdateUser(ctx, user.ID, UserParams{Email: fake.EmailAddress()})
		check.NoError(err)
		check.NotEqual(user, usr)

		users, err := bk.ListUsers(ctx, nil)
		check.NoError(err)
		// check.Contains(users, usr)
		check.GreaterOrEqual(len(users), 3)

		ok, err := bk.DeleteUser(ctx, usr.ID, nil)
		check.NoError(err)
		check.True(ok)

		// Create Book 1
		book, err := bk.CreateBook(ctx, BookParams{Name: fake.Title(), Description: fake.Words()})
		check.NoError(err)
		check.NotEmpty(book)

		// Create Book 2
		book2, err := bk.CreateBook(ctx, BookParams{Name: fake.Title(), Description: fake.Words(), Image: img})
		check.NoError(err)
		check.NotEmpty(book2)

		// List Books
		books, err := bk.ListBooks(ctx, nil)
		check.NoError(err)
		check.Len(books, 2)

		// Get Book 1
		book1, err := bk.GetBook(ctx, book.ID)
		check.NoError(err)
		check.Equal(book.Name, book1.Name)

		// Update Book 1
		book3, err := bk.UpdateBook(ctx, book.ID, BookParams{Name: fake.Title()})
		check.NoError(err)
		check.NotEqual(book3, book)

		// Delete Book 2
		ok, err = bk.DeleteBook(ctx, book2.ID)
		check.NoError(err)
		check.True(ok)

		// Check if book was delete
		books, err = bk.ListBooks(ctx, nil)
		check.NoError(err)
		check.Len(books, 1)

		// Create Shelf
		shelf, err := bk.CreateShelf(ctx, ShelfParams{
			Name:        fake.Title(),
			Description: fake.Words(),
			Books:       []int{book.ID},
		})
		check.NoError(err)

		// Update Shelf (Add Book)
		shelfU, err := bk.UpdateShelf(ctx, shelf.ID, ShelfParams{
			Name: fake.Title(),
		})
		check.NoError(err)
		check.NotEqual(shelf, shelfU)

		// Get Shelf
		shelf1, err := bk.GetShelf(ctx, shelf.ID)
		check.NoError(err)
		check.Equal(shelf1.Name, shelfU.Name)

		// List Shelves
		shelves, err := bk.ListShelves(ctx, nil)
		check.NoError(err)
		check.Len(shelves, 1)

		tagParams := []TagParams{
			{
				Name:  fake.Title(),
				Value: fake.WordsN(2),
			},
			{
				Name:  fake.Title(),
				Value: fake.WordsN(2),
			},
			{
				Name:  fake.Title(),
				Value: fake.WordsN(2),
			},
		}

		// Create Chapter
		chapter, err := bk.CreateChapter(ctx, ChapterParams{
			BookID: book.ID,
			Name:   fake.Title(),
			Tags:   tagParams,
		})

		check.NoError(err)
		check.NotEmpty(chapter)

		// Update Chapter
		chapterUpdated, err := bk.UpdateChapter(ctx, chapter.ID, ChapterParams{Name: fake.Title()})
		check.NoError(err)
		check.NotEqual(chapterUpdated.Name, chapter.Name)

		// List Chapter
		chapters, err := bk.ListChapters(ctx, nil)
		check.NoError(err)
		check.Len(chapters, 1)
		check.Contains(chapters, chapterUpdated)

		// Get Chapter
		detailedChapter, err := bk.GetChapter(ctx, chapter.ID)
		check.NoError(err)
		check.Equal(detailedChapter.Name, chapterUpdated.Name)
		check.Len(detailedChapter.Tags, len(tagParams))

		pageParams := PageParams{
			// ChapterID: chapter.ID,
			BookID: book.ID,
			Name:   fake.Title(),
			// Markdown:  fake.WordsN(rand.Intn(1000) + 10),
			HTML: fmt.Sprintf("<p>%s</p>", fake.WordsN(rand.Intn(1000)+10)),
			Tags: tagParams,
		}

		// Create Page
		page, err := bk.CreatePage(ctx, pageParams)
		check.NoError(err, "unable to create page %+v\n", pageParams)
		check.Equal(page.BookID, book.ID)

		pageUpdateParams := PageParams{
			Name: fake.Title(),
		}

		// Update Page
		pageUpdated, err := bk.UpdatePage(ctx, page.ID, pageUpdateParams)
		check.NoError(err)
		check.NotEqual(pageUpdated, page)

		// List Page
		pages, err := bk.ListPages(ctx, nil)
		check.NoError(err)
		check.Len(pages, 1)

		// Get Page
		pageDetailed, err := bk.GetPage(ctx, page.ID)
		check.NoError(err)
		check.NotEmpty(pageDetailed.HTML)
		check.Equal(pageDetailed.Markdown, pageParams.Markdown)

		// Attachments
		attachment, err := bk.CreateAttachment(ctx, AttachmentParams{
			Name:       fake.Title(),
			UploadedTo: page.ID,
			File:       img,
		})
		check.NoError(err)
		check.NotEmpty(attachment)

		// List Attachments
		attachments, err := bk.ListAttachments(ctx, nil)
		check.NoError(err)
		check.Len(attachments, 1)

		// Update attachment
		attachmentUpdated, err := bk.UpdateAttachment(ctx, attachment.ID, AttachmentParams{Name: fake.Title()})
		check.NoError(err)
		check.NotEqual(attachment, attachmentUpdated)

		// Detailed Attachment
		detailedAttachment, err := bk.GetAttachment(ctx, attachment.ID)
		check.NoError(err)
		check.NotEmpty(detailedAttachment)

		// Search
		results, err := bk.Search(ctx, SearchParams{InName: &pageUpdateParams.Name})
		check.NoError(err)
		check.Len(results, 1)

		check.Equal(ContentPage, results[0].Type)
		check.Equal(pageUpdateParams.Name, results[0].Name)

		results, err = bk.Search(ctx, SearchParams{InName: &pageUpdateParams.Name, Type: []ContentType{ContentBook}})
		check.NoError(err)
		check.Len(results, 0)

		// Delete Attachment
		ok, err = bk.DeleteAttachment(ctx, attachment.ID)
		check.NoError(err)
		check.True(ok)

		// Delete Page
		ok, err = bk.DeletePage(ctx, page.ID)
		check.NoError(err)
		check.True(ok)

		pages, err = bk.ListPages(ctx, nil)
		check.NoError(err)
		check.Len(pages, 0)

		// Delete Chapter
		ok, err = bk.DeleteChapter(ctx, chapter.ID)
		check.NoError(err)
		check.True(ok)

		chapters, err = bk.ListChapters(ctx, nil)
		check.NoError(err)
		check.Len(chapters, 0)

		// Delete Book 1
		ok, err = bk.DeleteBook(ctx, book.ID)
		check.NoError(err)
		check.True(ok)

		books, err = bk.ListBooks(ctx, nil)
		check.NoError(err)
		check.Len(books, 0)

		// Delete Shelves
		ok, err = bk.DeleteShelf(ctx, shelf.ID)
		check.NoError(err)
		check.True(ok)

		shelves, err = bk.ListShelves(ctx, nil)
		check.NoError(err)
		check.Len(shelves, 0)

		// List recycle bin items
		items, err := bk.ListRecycleBinItems(ctx)
		check.NoError(err)
		check.Len(items, 6)

		for _, item := range items {

			switch item.DeletableType {
			case ContentBook:
				item, ok := item.Book()
				check.True(ok)
				check.NotNil(item)
				check.NotEmpty(item)
			case ContentChapter:
				item, ok := item.Chapter()
				check.True(ok)
				check.NotNil(item)
				check.NotEmpty(item)
				check.NotEmpty(item.Parent)
			case ContentShelf:
				item, ok := item.Shelf()
				check.True(ok)
				check.NotNil(item)
				check.NotEmpty(item)
			case ContentPage:
				item, ok := item.Page()
				check.True(ok)
				check.NotNil(item)
				check.NotEmpty(item)
				check.NotEmpty(item.Parent)
			default:
				t.Errorf("unknown type %v", item.DeletableType)
			}

		}

		item := items[0]

		// Restore recycle bin item
		count, err := bk.RestoreRecyleBinItem(ctx, item.DeletableID)
		check.NoError(err)
		check.Equal(count, 1)

		items, err = bk.ListRecycleBinItems(ctx)
		check.NoError(err)
		check.Len(items, 5)

		item = items[0]

		// Destroy item
		count, err = bk.DeleteRecycleBinItem(ctx, item.DeletableID)
		check.NoError(err)
		check.Equal(count, 1)

		// List recycle bin items
		items, err = bk.ListRecycleBinItems(ctx)
		check.NoError(err)
		check.Len(items, 4)

	})

	// DELETE TOKEN
	check.NoError(page.Navigate(page.MustInfo().URL + "/delete"))
	check.NoError(page.MustElement("#main-content > div > div > div > div > form > button").Click(proto.InputMouseButtonLeft))

}
