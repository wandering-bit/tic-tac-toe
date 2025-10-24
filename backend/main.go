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
        logger.Info("Hello World!")
        nk.LeaderboardCreate(ctx, "leaderboard", true, "desc", "incr", "", map[string]interface{}{})
        initializer.RegisterRpc("match", rpcMatchFinder)
        initializer.RegisterRpc("leaderboard", rpcGetLeaderboard)
        initializer.RegisterMatch("tic-tac", start)

        return nil
}

func start(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
        logger.Debug("hi bro debug")
        logger.Info("hi bro info")
        logger.Error("hi bro error")
        return &MathHandler{}, nil
}

func rpcMatchFinder(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
        logger.Info("received match find request balfkasf")
        minSize := 1
        maxSize := 1
        logger.Info("getting match lists")
        matches, err := nk.MatchList(ctx, 1, true, "", &minSize, &maxSize, "")
        logger.Info("mathces are ", matches)
        matchIds := make([]string, 0)
        logger.Info("initialised ids list")
        if err != nil {
                logger.Error("Failed to fetch matches", err)
        }
        if len(matches) > 0 {
                logger.Info("match already exsitss")
                for _, match := range matches {
                        matchIds = append(matchIds, match.MatchId)
                }
        } else {
                logger.Info("now matches exist")
                matchId, err := nk.MatchCreate(ctx, "tic-tac", map[string]interface{}{})
                logger.Debug("generated id is ", matchId)
                if err != nil {
                        logger.Error("Failed to create match ", err)
                }
                matchIds = append(matchIds, matchId)
        }
        logger.Info("here")
        // jsonstring, _ := json.Marshal(matchIds)
        // logger.Debug("ids are ", jsonstring)
        return matchIds[0], nil
}
func rpcGetLeaderboard(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
        records, _, _, _, _ := nk.LeaderboardRecordsList(ctx, "leaderboard", nil, 3, "", 0)
        jsonString, _ := json.Marshal(&records)
        return string(jsonString), nil
}

// func (ctx context.Context, logger runtime.Logger,
//      db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
//      logger.Debug("hi bro debug")
//      logger.Info("hi bro info")
//      logger.Error("hi bro error")
//      return &LobbyMatch{}, nil
// }

func (m *MathHandler) MatchInit(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
        logger.Info("inside match init")
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
        logger.Info("match join attemp")
        // s := state.(*MatchState)
        // logger.Debug("state is ", s)
        return state, true, ""
}
func (m *MathHandler) MatchJoin(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64,
        state interface{}, presences []runtime.Presence) interface{} {
        logger.Debug("precences are ", presences)
        s := state.(*MatchState)
        for _, presence := range presences {
                s.PresencesMap[presence.GetUsername()] = presence
                s.Opponents = append(s.Opponents, presence.GetUsername())
                nk.LeaderboardRecordWrite(ctx, "leaderboard", presence.GetUserId(), presence.GetUsername(), 0, 0, nil, nil)
        }
        // s.PresencesMap[presences[0].GetUserId()] = presences[0]
        // s.Opponents = append(s.Opponents, presences[0].GetUsername())
        logger.Debug("inside match join")

        if len(s.PresencesMap) == 2 {
                time.Sleep(2 * time.Second)
                logger.Debug("received match")
                s.Turn = s.Opponents[0]
                s.Symbol[s.Opponents[0]] = "X"
                s.Symbol[s.Opponents[1]] = "O"
                jsonString, err := json.Marshal(s)
                if err != nil {
                        logger.Error("error sending ack ", err)
                }
                logger.Debug("sending packet ", jsonString)
                dispatcher.BroadcastMessage(1, []byte(jsonString), nil, nil, true)
        }
        return s
}
func (m *MathHandler) MatchLeave(ctx context.Context, logger runtime.Logger,
        db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher,
        tick int64, state interface{}, presences []runtime.Presence) interface{} {
        logger.Debug("match leave presence is ", presences)
        // for _, presence := range presences {
        //      delete()
        // }
        return nil
}
func (m *MathHandler) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB,
        nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{},
        messages []runtime.MatchData) interface{} {
        logger.Info("inside game loop")
        s := state.(*MatchState)
        logger.Info("after type cast")
        logger.Debug("message after cast is ", messages)
        if len(messages) > 0 {
                logger.Debug("messages length check ", messages[0].GetData())
        }
        if len(messages) > 0 && len(s.PresencesMap) == 2 {
                logger.Info("got message from clients")
                opCode := messages[0].GetOpCode()
                switch opCode {
                case 2:
                        data := []string{}
                        err := json.Unmarshal(messages[0].GetData(), &data)
                        if err != nil {
                                logger.Error("unmarshal error is ", err)
                        }
                        s.Positions = data
                        winning := checkWinner(s, logger)
                        if winning {
                                logger.Debug("inside winning")
                                s.Winner = s.Turn
                                jsonString, _ := json.Marshal(s)
                                dispatcher.BroadcastMessage(3, []byte(jsonString), nil, nil, true)
                                logger.Info("going to save to memory")
                                nk.LeaderboardRecordWrite(ctx, "leaderboard", s.PresencesMap[s.Winner].GetUserId(), s.Winner, 10, 0, nil, nil)
                                logger.Info("saved to leaderboard")
                                time.Sleep(5 * time.Second)
                                records, _, _, _, _ := nk.LeaderboardRecordsList(ctx, "leaderboard", nil, 10, "", 0)
                                logger.Info("records are", records)
                                return nil
                        }
                        logger.Debug("not inside winning")
                        logger.Debug("positions date is ", s.Positions)
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
                logger.Debug("solution is ", solution)
                winning = true
                for _, location := range solution {
                        logger.Debug("position is ", s.Positions)
                        logger.Debug("turn is ", s.Turn)
                        logger.Debug("location is ", location)
                        logger.Debug("Symbol is ", s.Symbol[s.Turn])
                        if s.Positions[location] != s.Symbol[s.Turn] {
                                logger.Debug("got false")
                                winning = false
                                break
                        }
                }
                if winning {
                        return winning
                }
        }
        logger.Debug("end result is ", winning)
        return winning
}

// func saveToMemory(ctx context.Context, nk runtime.NakamaModule, userId string, logger runtime.Logger) {
//      object, err := nk.StorageWrite(ctx, []*runtime.StorageWrite{{
//              Collection:      "LeaderBoard",
//              Key:             "Scores",
//              UserID:          userId,
//              Value:           "100",
//              Version:         "1",
//              PermissionRead:  1,
//              PermissionWrite: 1,
//      }})
//      if err != nil {
//              logger.Error("error saving to memory", err)
//      }
//      logger.Debug("save object ", object)
// }

// func deleteFromMemroy(ctx context.Context, nk runtime.NakamaModule) {}
// func readFromMemory(ctx context.Context, nk runtime.NakamaModule, userId string, logger runtime.Logger) {
//      objects, err := nk.StorageRead(ctx, []*runtime.StorageRead{{
//              Collection: "LeaderBoard",
//              Key:        "Scores",
//              UserID:     userId,
//      }})
//      if err != nil {
//              logger.Error("error in fetching memory ", err)
//      }
//      logger.Info("objects size is ", len(objects))
//      logger.Info("objects is ", objects)
// }

// func fetchLeaderboard(ctx context.Context) {

// }