package telegram

import (
	"context"
	"fmt"
	"sort"

	"stool-grabber/internal/domain"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
)

// ScrapeChannelComments собирает последние посты канала и комментарии через messages.getReplies.
//
// Ручной smoke (DoD Task 04): канал с включённой группой обсуждений; малый parse_depth;
// пост без комментариев не должен ломать сбор; delay_ms проверяется при необходимости замером времени.
func ScrapeChannelComments(ctx context.Context, api *tg.Client, params ScrapeParams) (*domain.ScrapeResult, error) {
	if params.ChannelUsername == "" {
		return nil, fmt.Errorf("channel username is empty")
	}

	username := NormalizeChannelUsername(params.ChannelUsername)

	resolved, err := invokeRPC(ctx, params.DelayMS, func(c context.Context) (*tg.ContactsResolvedPeer, error) {
		return api.ContactsResolveUsername(c, &tg.ContactsResolveUsernameRequest{Username: username})
	})
	if err != nil {
		return nil, fmt.Errorf("contacts.resolveUsername %q: %w", username, err)
	}

	pc, ok := resolved.Peer.(*tg.PeerChannel)
	if !ok {
		return nil, fmt.Errorf("peer for @%s is not a channel", username)
	}

	var channel *tg.Channel
	for _, chat := range resolved.Chats {
		ch, ok := chat.(*tg.Channel)
		if ok && ch.ID == pc.ChannelID {
			channel = ch
			break
		}
	}
	if channel == nil {
		return nil, fmt.Errorf("channel id %d not found in resolve response", pc.ChannelID)
	}

	inputChannel := &tg.InputChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	}
	inputPeer := &tg.InputPeerChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	}

	full, err := invokeRPC(ctx, params.DelayMS, func(c context.Context) (*tg.MessagesChatFull, error) {
		return api.ChannelsGetFullChannel(c, inputChannel)
	})
	if err != nil {
		return nil, fmt.Errorf("channels.getFullChannel: %w", err)
	}

	cFull, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, fmt.Errorf("channels.getFullChannel: expected channel full")
	}
	linkedID, linkedOK := cFull.GetLinkedChatID()
	if !linkedOK || linkedID == 0 {
		return nil, fmt.Errorf("канал @%s не связан с группой обсуждений (linked_chat_id)", username)
	}

	posts, err := fetchChannelPosts(ctx, api, inputPeer, channel.ID, params.ParseDepth, params.DelayMS)
	if err != nil {
		return nil, err
	}

	out := &domain.ScrapeResult{
		ChannelUsername:        username,
		LinkedDiscussionChatID: linkedID,
		ChannelAdminUserIDs:    nil,
		Threads:                make([]domain.PostThread, 0, len(posts)),
	}

	for _, msg := range posts {
		comments, err := fetchRepliesForPost(ctx, api, inputPeer, msg.ID, params.DelayMS)
		if err != nil {
			return nil, fmt.Errorf("getReplies for channel msg %d: %w", msg.ID, err)
		}
		out.Threads = append(out.Threads, domain.PostThread{
			ChannelMessageID: msg.ID,
			Comments:         comments,
		})
	}

	if params.ExcludeAdmins {
		adminIDs, err := fetchChannelAdminUserIDs(ctx, api, inputChannel, params.DelayMS)
		if err != nil {
			// For public channels Telegram may require admin rights to list admins.
			// In that case we continue without admin filtering.
			if tgerr.Is(err, "CHAT_ADMIN_REQUIRED") {
				adminIDs = nil
			} else {
				return nil, fmt.Errorf("fetch channel admins: %w", err)
			}
		}
		out.ChannelAdminUserIDs = adminIDs
	}

	return out, nil
}

func unwrapMessagesBox(box tg.MessagesMessagesClass) ([]tg.MessageClass, []tg.ChatClass, []tg.UserClass) {
	if box == nil {
		return nil, nil, nil
	}
	switch v := box.(type) {
	case *tg.MessagesMessages:
		return v.Messages, v.Chats, v.Users
	case *tg.MessagesMessagesSlice:
		return v.Messages, v.Chats, v.Users
	case *tg.MessagesChannelMessages:
		return v.Messages, v.Chats, v.Users
	case *tg.MessagesMessagesNotModified:
		return nil, nil, nil
	default:
		return nil, nil, nil
	}
}

func fetchChannelAdminUserIDs(ctx context.Context, api *tg.Client, inputChannel tg.InputChannelClass, delayMS int) ([]int64, error) {
	const page = 200
	offset := 0
	ids := make(map[int64]struct{})

	for {
		off := offset
		boxed, err := invokeRPC(ctx, delayMS, func(c context.Context) (tg.ChannelsChannelParticipantsClass, error) {
			return api.ChannelsGetParticipants(c, &tg.ChannelsGetParticipantsRequest{
				Channel: inputChannel,
				Filter:  &tg.ChannelParticipantsAdmins{},
				Offset:  off,
				Limit:   page,
				Hash:    0,
			})
		})
		if err != nil {
			return nil, err
		}
		mod, ok := boxed.AsModified()
		if !ok || mod == nil {
			break
		}
		if len(mod.Participants) == 0 {
			break
		}
		for _, p := range mod.Participants {
			switch v := p.(type) {
			case *tg.ChannelParticipantAdmin:
				ids[v.UserID] = struct{}{}
			case *tg.ChannelParticipantCreator:
				ids[v.UserID] = struct{}{}
			}
		}
		if len(mod.Participants) < page {
			break
		}
		offset += len(mod.Participants)
	}

	out := make([]int64, 0, len(ids))
	for id := range ids {
		if id == 0 {
			continue
		}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

func fetchChannelPosts(
	ctx context.Context,
	api *tg.Client,
	channelPeer tg.InputPeerClass,
	channelID int64,
	parseDepth int,
	delayMS int,
) ([]*tg.Message, error) {
	const page = 100
	offsetID := 0
	seen := make(map[int]struct{})
	posts := make([]*tg.Message, 0, parseDepth)

	for len(posts) < parseDepth {
		off := offsetID
		boxed, err := invokeRPC(ctx, delayMS, func(c context.Context) (tg.MessagesMessagesClass, error) {
			return api.MessagesGetHistory(c, &tg.MessagesGetHistoryRequest{
				Peer:       channelPeer,
				Limit:      page,
				OffsetID:   off,
				OffsetDate: 0,
				AddOffset:  0,
				MaxID:      0,
				MinID:      0,
				Hash:       0,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("messages.getHistory: %w", err)
		}
		rawMsgs, _, _ := unwrapMessagesBox(boxed)
		if len(rawMsgs) == 0 {
			break
		}

		minAcross := 0
		novel := 0
		for _, mc := range rawMsgs {
			msg, ok := mc.(*tg.Message)
			if !ok || !isChannelPost(msg, channelID) {
				continue
			}
			if _, dup := seen[msg.ID]; dup {
				continue
			}
			seen[msg.ID] = struct{}{}
			novel++
			posts = append(posts, msg)
			if len(posts) >= parseDepth {
				break
			}
			if minAcross == 0 || msg.ID < minAcross {
				minAcross = msg.ID
			}
		}
		if novel == 0 {
			break
		}
		if len(posts) >= parseDepth {
			break
		}
		if minAcross == 0 {
			break
		}
		offsetID = minAcross
	}
	return posts, nil
}

func isChannelPost(m *tg.Message, channelID int64) bool {
	if m == nil || m.PeerID == nil {
		return false
	}
	ch, ok := m.PeerID.(*tg.PeerChannel)
	return ok && ch.ChannelID == channelID
}

func fetchRepliesForPost(
	ctx context.Context,
	api *tg.Client,
	channelPeer tg.InputPeerClass,
	channelPostID int,
	delayMS int,
) ([]domain.Comment, error) {
	const page = 100
	offsetID := 0
	offsetDate := 0
	seen := make(map[int]struct{})
	var comments []domain.Comment

	for {
		oID, oDate := offsetID, offsetDate
		boxed, err := invokeRPC(ctx, delayMS, func(c context.Context) (tg.MessagesMessagesClass, error) {
			return api.MessagesGetReplies(c, &tg.MessagesGetRepliesRequest{
				Peer:       channelPeer,
				MsgID:      channelPostID,
				OffsetID:   oID,
				OffsetDate: oDate,
				AddOffset:  0,
				Limit:      page,
				MaxID:      0,
				MinID:      0,
				Hash:       0,
			})
		})
		if err != nil {
			return nil, err
		}
		rawMsgs, _, _ := unwrapMessagesBox(boxed)
		if len(rawMsgs) == 0 {
			break
		}

		novel := 0
		nextMinID := 0
		nextDate := 0
		for _, mc := range rawMsgs {
			msg, ok := mc.(*tg.Message)
			if !ok {
				continue
			}
			if _, dup := seen[msg.ID]; dup {
				continue
			}
			seen[msg.ID] = struct{}{}
			novel++
			comments = append(comments, rawCommentFromMessage(msg))

			if nextMinID == 0 || msg.ID < nextMinID {
				nextMinID = msg.ID
				nextDate = msg.Date
			}
		}
		if novel == 0 {
			break
		}
		offsetID = nextMinID
		offsetDate = nextDate
	}
	return comments, nil
}

func rawCommentFromMessage(m *tg.Message) domain.Comment {
	var uid int64
	if m.FromID != nil {
		if u, ok := m.FromID.(*tg.PeerUser); ok {
			uid = u.UserID
		}
	}
	return domain.Comment{
		MessageID:    m.ID,
		SenderUserID: uid,
		Text:         m.Message,
		DateUnix:     m.Date,
	}
}
