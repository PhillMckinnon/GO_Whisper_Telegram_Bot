package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type userState struct {
	Action    string
	HasMedia  bool
	MediaFile string
	TextInput string
	Language  string
}

var userStates = make(map[int64]*userState)
var languages = []telego.InlineKeyboardButton{
	tu.InlineKeyboardButton("English").WithCallbackData("lang_en"),
	tu.InlineKeyboardButton("Arabic").WithCallbackData("lang_ar"),
	tu.InlineKeyboardButton("Czech").WithCallbackData("lang_cz"),
	tu.InlineKeyboardButton("Dutch").WithCallbackData("lang_nl"),
	tu.InlineKeyboardButton("Deutsch").WithCallbackData("lang_de"),
	tu.InlineKeyboardButton("Français").WithCallbackData("lang_fr"),
	tu.InlineKeyboardButton("Español").WithCallbackData("lang_es"),
	tu.InlineKeyboardButton("Italiano").WithCallbackData("lang_it"),
	tu.InlineKeyboardButton("Portuguese").WithCallbackData("lang_pt"),
	tu.InlineKeyboardButton("Polish").WithCallbackData("lang_pl"),
	tu.InlineKeyboardButton("Turkish").WithCallbackData("lang_tr"),
	tu.InlineKeyboardButton("Hungarian").WithCallbackData("lang_hu"),
	tu.InlineKeyboardButton("Russian").WithCallbackData("lang_ru"),
	tu.InlineKeyboardButton("Chinese").WithCallbackData("lang_zh"),
}

func DotEnvVar(key string) string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	return os.Getenv(key)
}
func main() {
	ctx := context.Background()
	// Get Bot token from environment variables
	botToken := DotEnvVar("TOKEN")
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatalf("failed to create bot: %s", err)
	}

	//start polling
	upd, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		log.Fatalf("failed to get updates: %s", err)
	}
	bh, _ := th.NewBotHandler(bot, upd)
	defer func() { _ = bh.Stop() }()

	for update := range upd {
		switch {
		case update.CallbackQuery != nil:

			cb := update.CallbackQuery
			langCode := strings.TrimPrefix(cb.Data, "lang_")
			state := userStates[cb.From.ID]
			if state.Action == "await_language" {
				state.Language = langCode
				go bot.AnswerCallbackQuery(context.Background(), &telego.AnswerCallbackQueryParams{
					CallbackQueryID: cb.ID,
					Text:            "Language set to: " + langCode,
				})
				msg := cb.Message.Message()
				if msg != nil {
					chatID := msg.Chat.ID
					sendMessage(bot, chatID, "Processing...")
					//send to backend
					err := sendToBackend(bot, chatID, state, "synthesize")
					if err != nil {
						sendMessage(bot, chatID, "Error sending to backend: "+err.Error())
					}
				}
				state.Action = ""

			}

		case update.Message != nil:
			msg := update.Message
			userID := msg.Chat.ID
			nme := update.Message.From.Username

			// Initialize user state
			if _, exists := userStates[userID]; !exists {
				userStates[userID] = &userState{}
			}
			state := userStates[userID]

			// Check for commands or keyboard inputs
			switch {
			case msg.Text == "Transcribe":
				state.Action = "transcribe"
				state.HasMedia = false
				sendMessage(bot, userID, "Now, send a file (mp4, mp3, ogg, wav).")
			case msg.Text == "Synthesize":
				state.Action = "synthesize"
				state.HasMedia = false
				state.TextInput = ""
				sendMessage(bot, userID, "Now, send a file (mp4, mp3, ogg, wav).")
			case msg.Document != nil || msg.Audio != nil || msg.Video != nil || msg.Voice != nil:
				if state.Action == "" {
					messageText := fmt.Sprintf("%s, please choose an action:", nme)
					sendWithKeyboard(bot, userID, messageText)
					continue
				}
				state.HasMedia = true
				state.MediaFile = getFileID(msg)
				if state.Action == "transcribe" {
					sendMessage(bot, userID, "Please wait...")
					err := sendToBackend(bot, userID, state, string(state.Action))
					if err != nil {
						sendMessage(bot, userID, "Error sending to backend: "+err.Error())
					}
					state.Action = ""

				} else if state.Action == "synthesize" {
					sendMessage(bot, userID, "Now send the text to synthesize.")
				}
			case state.Action == "synthesize" && state.HasMedia && msg.Text != "":
				state.TextInput = msg.Text
				sendLanguageSelection(bot, userID)
				state.Action = "await_language"
			default:
				messageText := fmt.Sprintf("%s, please choose an action:", nme)
				sendWithKeyboard(bot, userID, messageText)
			}
		}
	}
}

func getFileExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext != "" {
		return ext[1:]
	}
	return ".bin"
}
func sendMessage(bot *telego.Bot, chatID int64, text string) {
	_, _ = bot.SendMessage(context.Background(), tu.Message(tu.ID(chatID), text))
}
func getFileDownloadURL(bot *telego.Bot, fileID string) (string, error) {
	fileResp, err := bot.GetFile(context.Background(), &telego.GetFileParams{FileID: fileID})
	if err != nil {
		return "", err
	}
	return "https://api.telegram.org/file/bot" + bot.Token() + "/" + fileResp.FilePath, nil
}
func sendWithKeyboard(bot *telego.Bot, chatID int64, text string) {
	keyboard := tu.Keyboard(
		tu.KeyboardRow(
			tu.KeyboardButton("Transcribe"),
			tu.KeyboardButton("Synthesize"),
		),
	).WithResizeKeyboard().WithInputFieldPlaceholder("Select one")
	_, _ = bot.SendMessage(context.Background(), tu.Message(tu.ID(chatID), text).WithReplyMarkup(keyboard))
}
func sendLanguageSelection(bot *telego.Bot, chatID int64) {
	keyboard := &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			languages[:4],
			languages[4:8],
			languages[8:],
		},
	}
	_, _ = bot.SendMessage(context.Background(), tu.Message(tu.ID(chatID), "Choose output language (default is English):").WithReplyMarkup(keyboard))
}
func sendToBackend(bot *telego.Bot, chatID int64, state *userState, realAction string) error {
	textInput := state.TextInput
	language := state.Language
	fileResp, err := bot.GetFile(context.Background(), &telego.GetFileParams{FileID: state.MediaFile})
	if err != nil {
		return fmt.Errorf("failed to get file: %w", err)
	}
	fileURL, err := getFileDownloadURL(bot, state.MediaFile)
	if err != nil {
		return fmt.Errorf("failed to get file URL: %w", err)
	}
	resp, err := http.Get(fileURL)

	if err != nil {
		return fmt.Errorf("failed to download media file: %w", err)
	}
	defer resp.Body.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", "media."+getFileExtension(fileResp.FilePath))
	fmt.Println("FilePath:", fileResp.FilePath)
	ext := getFileExtension(fileResp.FilePath)
	fmt.Println("Extension:", ext)
	if err != nil {
		return fmt.Errorf("failed to create file part: %w", err)
	}
	if _, err := io.Copy(part, resp.Body); err != nil {
		return fmt.Errorf("failed to copy media file: %w", err)
	}
	if textInput != "" {
		_ = writer.WriteField("text", textInput)
	}
	if language != "" {
		_ = writer.WriteField("language", language)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	backendURL := "http://backend:5000/api/synthesize"
	if realAction == "transcribe" {
		backendURL = "http://backend:5000/api/transcribe"
	}

	req, err := http.NewRequest("POST", backendURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 720 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to backend: %w", err)
	}
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read backend response: %w", err)
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf("backend returned error: %s", res.Status)
	}
	fmt.Println("state.Action:", state.Action)
	switch realAction {
	case "transcribe":
		var parsed struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(bodyBytes, &parsed); err != nil || parsed.Text == "" {
			return fmt.Errorf("invalid transcribe response: %w", err)
		}
		sendMessage(bot, chatID, "Transcription:\n"+parsed.Text)
	case "synthesize":
		fmt.Println("Sending .wav to Telegram")
		fmt.Printf("Received audio file size: %d bytes\n", len(bodyBytes))
		fmt.Printf("Response Content-Type: %s\n", res.Header.Get("Content-Type"))
		if len(bodyBytes) < 100 {
			fmt.Println("Warning: Audio body seems suspiciously small")
		}

		_, err = bot.SendAudio(context.Background(), &telego.SendAudioParams{ChatID: tu.ID(chatID), Audio: tu.FileFromReader(bytes.NewReader(bodyBytes), "result.wav")})
		if err != nil {
			return fmt.Errorf("failed to send audio: %w", err)
		}
	default:
		fmt.Println("Unexpected action value:", state.Action)
	}
	return nil
}
func getFileID(msg *telego.Message) string {
	if msg.Document != nil {
		return msg.Document.FileID
	}
	if msg.Audio != nil {
		return msg.Audio.FileID
	}
	if msg.Video != nil {
		return msg.Video.FileID
	}
	if msg.Voice != nil {
		return msg.Voice.FileID
	}
	return ""
}
