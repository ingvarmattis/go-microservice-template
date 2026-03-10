package server

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"
)

// TelegramBotInterface is the contract for telegram bot (real or noop).
// Use NotifyMessage for generic text notifications (e.g. from your domain handlers).
type TelegramBotInterface interface {
	NotifyMessage(msg string)
	Start()
	Close()
}

type NoopTelegramBot struct{}

func (NoopTelegramBot) NotifyMessage(string) {}
func (NoopTelegramBot) Start()               {}
func (NoopTelegramBot) Close()               {}

// NewNoopTelegramBot returns a no-op telegram bot.
func NewNoopTelegramBot() *NoopTelegramBot {
	return &NoopTelegramBot{}
}

type TelegramBot struct {
	tb             *telebot.Bot
	logger         *zap.Logger
	allowedChatIDs []int64
}

// NewTelegramBot creates a real Telegram bot. Use NotifyMessage to send text to all allowed chats.
func NewTelegramBot(
	logger *zap.Logger, token string, timeout time.Duration, allowedChatIDs []int64,
) (*TelegramBot, error) {
	if token == "" || len(allowedChatIDs) == 0 {
		return nil, fmt.Errorf("telegram: token and allowed_chat_ids required when enabled")
	}

	pref := telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: timeout},
	}

	tBot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot | %w", err)
	}

	tBot.Use(middleware.Whitelist(allowedChatIDs...))

	bot := &TelegramBot{
		tb:             tBot,
		logger:         logger,
		allowedChatIDs: allowedChatIDs,
	}

	tBot.Use(bot.loggingMiddleware())

	tBot.Handle("/start", bot.onStart)

	return bot, nil
}

func (b *TelegramBot) loggingMiddleware() func(telebot.HandlerFunc) telebot.HandlerFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			chatID := c.Chat().ID
			var command string
			if c.Message() != nil && c.Message().Text != "" && c.Message().Text[0] == '/' {
				parts := strings.Fields(c.Message().Text)
				if len(parts) > 0 {
					command = parts[0]
				}
			}

			b.logger.Info("telegram request",
				zap.Int64("chat_id", chatID),
				zap.Duration("duration", duration),
				zap.String("command", command),
				zap.Time("time", start),
			)
			return err
		}
	}
}

func (b *TelegramBot) onStart(ctx telebot.Context) error {
	return ctx.Send("You have subscribed to updates.")
}

func (b *TelegramBot) Start() {
	b.logger.Info("starting telegram bot")
	b.tb.Start()
}

func (b *TelegramBot) Close() {
	b.tb.Stop()
	_, _ = b.tb.Close()
}

// NotifyMessage sends the given text to all allowed chats (HTML parse mode).
func (b *TelegramBot) NotifyMessage(msg string) {
	if msg == "" {
		return
	}
	opts := &telebot.SendOptions{ParseMode: telebot.ModeHTML}
	for _, chatID := range b.allowedChatIDs {
		recipient := &telebot.Chat{ID: chatID}
		if _, err := b.tb.Send(recipient, msg, opts); err != nil {
			b.logger.Error("send telegram message", zap.Error(err), zap.Int64("chat_id", chatID))
		}
	}
}
