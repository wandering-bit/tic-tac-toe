import { Client } from './nakama-js.esm.mjs';

var useSSL = false;
var client = new Client("defaultkey", "15.206.174.20", "7350", useSSL);

var socket=null;
var opponentUserName = null;
var userName = null;
var currentSymbol = "";
var matchId = "";
document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
document.getElementById('home').classList.add('active')
const cells = document.querySelectorAll(".cell");

var positions = null;

initializeBoard();

const wins = [
    [0,1,2],
    [3,4,5],
    [6,7,8],
    [0,3,6],
    [1,4,7],
    [2,5,8],
    [0,4,8],
    [2,4,6]
];

const homeButton = document.getElementById("homecontinue");
const loginButton = document.getElementById("logincontinue");
const seeLeaderboardButton = document.getElementById("seeleaderboard");
const rematchButton = document.getElementById("rematchbutton")
const cancelButton = document.getElementById("cancelButton");

const loginInput = document.getElementById("inputusername");
const loginCross = document.getElementById("logincross");

const homePage = document.getElementById("home");
const loginPage = document.getElementById("login");
const matchPage = document.getElementById("match");
const gamePage = document.getElementById("game");
const resultPage = document.getElementById("resultPage")
const leaderboardPage = document.getElementById("leaderboard")

const myUsernameText = document.getElementById("playera");
const opponentUserNameText = document.getElementById("playerb");
const resultText = document.getElementById("result");

const turnText = document.getElementById("turntext");
const turnIcon = document.getElementById("turnicon")


var allowed= false;
var turn="";

homeButton.addEventListener('click', () => {
    homePage.classList.remove('active');
    loginPage.classList.add('active');
});

cancelButton.addEventListener('click', () => {
    socket.sendMatchState(matchId,4,JSON.stringify(userName));
    matchPage.classList.remove('active');
    leaderboardPage.classList.add('active');
    showLeaderboardPage(null,true);
    resultText.textContent = "";
});

rematchButton.addEventListener('click', async () => {
    leaderboardPage.classList.remove('active');
    matchPage.classList.add('active');
    positions=null;
    opponentUserName = null;
    currentSymbol = "";
    matchId = "";
    resetBoard();
    const ticket = await initMatchMaker();
    const details = await joinMatch(ticket);
    matchId = details.match_id;
});
loginButton.addEventListener('click', async () => {
    const username = loginInput.value;
    await getSession(username);
    loginPage.classList.remove('active');
    matchPage.classList.add('active');
    const ticket = await initMatchMaker();
    const details = await joinMatch(ticket);
    matchId = details.match_id;
    userName = username
    socket.onmatchdata = (matchstate) => {
        const jsonMatchState = StringToJson(matchstate.data);
        switch(matchstate.op_code){
                case 1:
                    initMatch(jsonMatchState);
                    break;
                case 2:
                    HandleGameMove(jsonMatchState);
                    break;
                case 3:
                    showLeaderboardPage(jsonMatchState,false);
                    break;
            }
    };
});
async function getSession(username){
    const create = true;
    const session = await client.authenticateCustom(
        generateId(),
        create,
        username
    );
    socket = client.createSocket(false);
    var appearOnline = true;
    await socket.connect(session, appearOnline);
}
function generateId() {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-";
    let id = "";
    for (let i = 0; i < 10; i++) {
    id += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return id;
}
async function initMatchMaker(){
    const ticket = await socket.rpc("match");
    return ticket;
}
async function joinMatch(ticket){
    const som = await socket.joinMatch(ticket.payload);
    return som;
}
function findOpponent(matchState){
    if(matchState.Opponents[0]==userName){
        opponentUserName = matchState.Opponents[1];
    }else{
        opponentUserName = matchState.Opponents[0]
    }
}
function initializeBoard(){
    cells.forEach(cell => cell.addEventListener("click", cellClicked));
}
function cellClicked(){
    const cellIndex = this.getAttribute("cellIndex");
    if (positions[cellIndex]!="" || !allowed){
        return;
    }
    updateCell(this,cellIndex);
    socket.sendMatchState(matchId,2,JSON.stringify(positions))
}
function updateCell(cell,index){
    positions[index]=currentSymbol;
    cell.textContent = currentSymbol;
}
function StringToJson(stringData){
    const data = new TextDecoder().decode(stringData);
    return JSON.parse(data);
}
function updateBoard(){
    cells.forEach(cell => {
        const cellIndex = cell.getAttribute("cellIndex");
        cell.textContent = positions[cellIndex];
    });
}
function resetBoard(){
    cells.forEach(cell => {
        cell.textContent = "";
    });
}
async function getLeaderboard(){
    const leaderboard = await socket.rpc("leaderboard")
    return JSON.parse(leaderboard.payload)
}

async function showLeaderboardPage(jsonMatchState,skip){
    if(!skip){
        gamePage.classList.remove('active');
        leaderboardPage.classList.add('active');
        if(jsonMatchState.Winner!=userName){
            resultText.textContent = "You Lost";
        }else{
            resultText.textContent = "You Won";
        }
    }
    const leaderboard = await getLeaderboard();
    leaderboard.forEach((element,index) => {
        const usercontainer = document.getElementById(("user"+index))
        const scorecontainer = document.getElementById(("points"+index))
        usercontainer.textContent = element.username.value
        const score = element.score ?? 0;
        scorecontainer.textContent = score
    });
}

async function initMatch(jsonMatchState){
    findOpponent(jsonMatchState);
    opponentUserNameText.textContent = opponentUserName;
    myUsernameText.textContent = userName;
    turn = jsonMatchState.Turn
    currentSymbol = jsonMatchState.Symbol[userName];
    positions = jsonMatchState.Positions;
    turnIcon.textContent = jsonMatchState.Symbol[userName]
    if(turn == userName){
        turnText.textContent = "YOUR TURN";
        allowed=true;
    }else{
        allowed=false;
        turnText.textContent = "OPPONENT'S TURN";
    }
    matchPage.classList.remove('active');
    gamePage.classList.add('active');
    // initializeBoard();
}

function HandleGameMove(jsonMatchState){
    turn = jsonMatchState.Turn
    positions = jsonMatchState.Positions;
    updateBoard();
    if(turn == userName){
        turnText.textContent = "YOUR TURN";
        allowed=true;
    }else{
        allowed=false;
        turnText.textContent = "OPPONENT'S TURN";
    }
}