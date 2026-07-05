package telegramtransport

import "time"

type broadcastDraft struct {
	Text      string
	Audience  string
	CreatedAt time.Time
}

func (b *Bot) setBroadcastDraft(adminID int64, draft broadcastDraft) {
	b.broadcastMu.Lock()
	defer b.broadcastMu.Unlock()
	b.broadcasts[adminID] = draft
}

func (b *Bot) getBroadcastDraft(adminID int64) (broadcastDraft, bool) {
	b.broadcastMu.RLock()
	defer b.broadcastMu.RUnlock()
	draft, ok := b.broadcasts[adminID]
	return draft, ok
}

func (b *Bot) clearBroadcastDraft(adminID int64) {
	b.broadcastMu.Lock()
	defer b.broadcastMu.Unlock()
	delete(b.broadcasts, adminID)
}
