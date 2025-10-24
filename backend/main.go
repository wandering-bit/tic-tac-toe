package main

import (
        "context"
        "database/sql"
        "encoding/json"
        "time"

        "github.com/heroiclabs/nakama-common/runtime"
)

type MathHandler struct{}

type MatchState struct {
        EmptyTicks     int
        PresencesMap   map[string]runtime.Presence
        Started        bool
        Turn           string
        Opponents      []string
        Symbol         map[string]string
        Positions      []string
        Winner         string
        RemainingUsers map[string]bool
}

var wins = [][]int{
        {0, 1, 2},
        {3, 4, 5},
        {6, 7, 8},
        {0, 3, 6},
        {1, 4, 7},
        {2, 5, 8},
        {0, 4, 8},
        {2, 4, 6},
}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
        nk.LeaderboardCreate(ctx, "leaderboard", true, "desc", "incr", "", map[string]interface{}{})
        initializer.RegisterRpc("match", rpcMatchFinder)
        initializer.RegisterRpc("leaderboard", rpcGetLeaderboard)
        initializer.RegisterMatch("tic-tac", start)

        return nil
}

func start(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
        return &MathHandler{}, nil
}

func rpcMatchFinder(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
        minSize := 1
        maxSize := 1
        matchIds := make([]string, 0)
        matches, err := nk.MatchList(ctx, 1, true, "", &minSize, &maxSize, "")
        if err != nil {
                logger.Error("Failed to fetch matches", err)
        }
        if len(matches) > 0 {
                for _, match := range matches {
                        matchIds = append(matchIds, match.MatchId)
                }
        } else {
                matchId, err := nk.MatchCreate(ctx, "tic-tac", map[string]interface{}{})
                if err != nil {
                        logger.Error("Failed to create match ", err)
                }
                matchIds = append(matchIds, matchId)
        }
        return matchIds[0], nil
}
func rpcGetLeaderboard(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
        records, _, _, _, _ := nk.LeaderboardRecordsList(ctx, "leaderboard", nil, 3, "", 0)
        jsonString, _ := json.Marshal(&records)
        return string(jsonString), nil
}

func (m *MathHandler) MatchInit(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
        state := &MatchState{
                EmptyTicks:     0,
                PresencesMap:   make(map[string]runtime.Presence, 2),
                Started:        false,
                Turn:           "",
                Opponents:      make([]string, 0),
                Symbol:         make(map[string]string, 0),
                Positions:      []string{"", "", "", "", "", "", "", "", ""},
                Winner:         "",
                RemainingUsers: make(map[string]bool, 0),
        }
        tickRate := 1
        label := ""
        return state, tickRate, label
}
func (m *MathHandler) MatchJoinAttempt(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{},
        presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
        return state, true, ""
}
func (m *MathHandler) MatchJoin(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64,
        state interface{}, presences []runtime.Presence) interface{} {
        s := state.(*MatchState)
        for _, presence := range presences {
                s.PresencesMap[presence.GetUsername()] = presence
                s.Opponents = append(s.Opponents, presence.GetUsername())
                nk.LeaderboardRecordWrite(ctx, "leaderboard", presence.GetUserId(), presence.GetUsername(), 0, 0, nil, nil)
        }

        if len(s.PresencesMap) == 2 {
                time.Sleep(2 * time.Second)
                s.Turn = s.Opponents[0]
                s.Symbol[s.Opponents[0]] = "X"
                s.Symbol[s.Opponents[1]] = "O"
                jsonString, err := json.Marshal(s)
                if err != nil {
                        logger.Error("error sending ack ", err)
                }
                dispatcher.BroadcastMessage(1, []byte(jsonString), nil, nil, true)
        }
        return s
}
func (m *MathHandler) MatchLeave(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher,
        tick int64, state interface{}, presences []runtime.Presence) interface{} {
        return nil
}
func (m *MathHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB,
        nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{},
        messages []runtime.MatchData) interface{} {
        s := state.(*MatchState)
        if len(messages) > 0{
                opCode := messages[0].GetOpCode()
                switch opCode {
                case 4:
                    data := ""
                    json.Unmarshal(messages[0].GetData(), &data)
                    presenceList := make([]runtime.Presence,0)
                    presenceList = append(presenceList,s.PresencesMap[data])
                    dispatcher.MatchKick(presenceList)
                    break;
                case 2:
                        data := []string{}
                        err := json.Unmarshal(messages[0].GetData(), &data)
                        if err != nil {
                                logger.Error("unmarshal error is ", err)
                        }
                        s.Positions = data
                        winning := checkWinner(s, logger)
                        if winning {
                                s.Winner = s.Turn
                                jsonString, _ := json.Marshal(s)
                                dispatcher.BroadcastMessage(3, []byte(jsonString), nil, nil, true)
                                nk.LeaderboardRecordWrite(ctx, "leaderboard", s.PresencesMap[s.Winner].GetUserId(), s.Winner, 10, 0, nil, nil)
                                return nil
                        }
                        if s.Turn == s.Opponents[0] {
                                s.Turn = s.Opponents[1]
                        } else {
                                s.Turn = s.Opponents[0]
                        }
                 }
                jsonString, err := json.Marshal(s)
                if err != nil {
                        logger.Error("error marshaling ", err)
                }
                dispatcher.BroadcastMessage(2, []byte(jsonString), nil, nil, true)
        }
        return state
}
func (m *MathHandler) MatchTerminate(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher,
        tick int64, state interface{}, graceSeconds int) interface{} {

        return nil
}
func (m *MathHandler) MatchSignal(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher,
        tick int64, state interface{}, data string) (interface{}, string) {
        return state, ""
}
func checkWinner(s *MatchState, logger runtime.Logger) bool {
        winning := true
        for _, solution := range wins {
                winning = true
                for _, location := range solution {
                        if s.Positions[location] != s.Symbol[s.Turn] {
                                winning = false
                                break
                        }
                }
                if winning {
                        return winning
                }
        }
        return winning
}