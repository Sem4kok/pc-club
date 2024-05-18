package main

import (
	"bufio"
	"container/list"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	ClubNotOpen    = "NotOpenYet"
	AlreadyInClub  = "YouShallNotPass"
	PlaceIsBusy    = "PlaceIsBusy"
	ClientNotExist = "ClientUnknown"
	CantWait       = "ICanWaitNoLonger!"
	NotSitting     = 0
)

type Event struct {
	Time   *Time
	ID     int
	Client string
	Table  int
}

type Table struct {
	IsBusy   bool
	Payments int
	WorkTime *Time
	LastSit  *Time
}

func main() {
	const op = "main.main"
	file := fileMustOpen()
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Printf("error closing file: %v", err)
		}
	}()

	writer := bufio.NewWriter(os.Stdout)
	scanner := bufio.NewScanner(file)

	// Here we only read the first three lines.
	// Queries will be processed later.
	numberOfTables, clubOpen, clubClose, costPerHour := readInitialParametersMust(scanner)

	// list of tables
	// 0th table does not exist
	// we must put tables into slice
	tables := make([]*Table, numberOfTables+1)
	for i := 0; i < numberOfTables+1; i++ {
		tables[i] = &Table{
			IsBusy:   false,
			Payments: 0,
			WorkTime: &Time{0, 0},
			LastSit:  nil,
		}
	}

	// queue of events
	queue := New()

	// to store is client in the club
	clientMap := make(map[string]int)

	if _, err := writer.WriteString(makeStringFromTime(clubOpen) + "\n"); err != nil {
		log.Fatal(fmt.Sprintf("%s : %v", op, err.Error()))
	}

	prevEventTime := &Time{
		Hour:   0,
		Minute: 0,
	}
	// start parse queries
	for scanner.Scan() {
		eventLine := scanner.Text()
		event := mustParseEvent(eventLine)

		// client can't come after closing in this task
		if isFirstTimeEarlier(clubClose, event.Time) {
			log.Fatal(eventLine)
		}

		// compare with previous event time
		if isFirstTimeEarlier(event.Time, prevEventTime) {
			log.Fatal(eventLine)
		}
		prevEventTime = event.Time

		// check if ID == 2, does table exist
		if event.ID == 2 {
			if event.Table == 0 || event.Table > numberOfTables {
				log.Fatal(eventLine)
			}
		}

		switch event.ID {
		case 1:
			playEvent1(&Event1{
				Event:     event,
				clubOpen:  clubOpen,
				clubClose: clubClose,
				ClientMap: &clientMap,
			}, writer)
		case 2:
			playEvent2(&Event2{
				Event:     event,
				Tables:    &tables,
				ClientMap: &clientMap,
			}, writer)
		case 3:
			playEvent3(&Event3{
				Event:     event,
				Tables:    &tables,
				ClientMap: &clientMap,
				Queue:     queue,
			}, writer)
		case 4:
			playEvent4(&Event4{
				Event:     event,
				Tables:    &tables,
				ClientMap: &clientMap,
				Queue:     queue,
			}, writer)
		}
	}

	// close club

	// first: sort clients, inside club by their name
	names := make([]string, 0, len(clientMap))
	for name := range clientMap {
		names = append(names, name)
	}
	sort.Strings(names)

	// play closing
	for _, name := range names {
		event := &Event{
			Time:   clubClose,
			ID:     4,
			Client: name,
		}
		closeEvent(&Event4{
			Event:     event,
			Tables:    &tables,
			ClientMap: &clientMap,
			Queue:     queue,
		}, writer)
	}
	timeCloseString := makeStringFromTime(clubClose)
	_, _ = writer.WriteString(timeCloseString + "\n")

	// print computing
	// for every table
	for i := 1; i < numberOfTables+1; i++ {
		table := tables[i]
		totalMoney := table.Payments * costPerHour
		timeString := makeStringFromTime(table.WorkTime)

		_, _ = writer.WriteString(fmt.Sprintf("%d %d ", i, totalMoney) + timeString + "\n")
	}

	_ = writer.Flush()
}

func fileMustOpen() *os.File {
	const op = "main.openFile"
	if args := os.Args; len(args) < 2 {
		log.Fatal(fmt.Sprintf("%s : file is required", op))
	}
	filename := os.Args[1]

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(fmt.Sprintf("%s : %s", op, err.Error()))
	}

	return file
}

func readInitialParametersMust(scanner *bufio.Scanner) (int, *Time, *Time, int) {
	var (
		numberOfTables      int
		clubOpen, clubClose *Time
		costPerHour         int
	)

	scanner.Scan()
	line1 := scanner.Text()
	numberOfTables, err := strconv.Atoi(line1)
	if err != nil || numberOfTables < 0 {
		log.Fatal(line1)
	}

	scanner.Scan()
	line2 := scanner.Text()
	times := strings.Split(line2, " ")
	if len(times) > 2 || len(times) < 2 {
		log.Fatal(line2)
	}

	clubOpen, err = parseTime(times[0])
	if err != nil {
		log.Fatal(line2)
	}

	clubClose, err = parseTime(times[1])
	if err != nil {
		log.Fatal(line2)
	}

	scanner.Scan()
	line3 := scanner.Text()
	costPerHour, err = strconv.Atoi(line3)
	if err != nil {
		log.Fatal(line3)
	}

	return numberOfTables, clubOpen, clubClose, costPerHour
}

// parseTime return Time struct in a valid format
// parseTime returns nil, if something gone wrong
func parseTime(timeString string) (*Time, error) {
	const op = "main.parseTime"
	// "hh:mm"
	if len(timeString) != 5 {
		return nil, fmt.Errorf(op + ": time has not valid format")
	}

	hour1, err := strconv.Atoi(string(timeString[0]))
	if err != nil {
		return nil, fmt.Errorf(op + ": hour must be a number")
	}

	hour2, err := strconv.Atoi(string(timeString[1]))
	if err != nil {
		return nil, fmt.Errorf(op + " : hour must be a number")
	}

	hour := hour1*10 + hour2
	if hour < 0 || hour > 23 {
		return nil, fmt.Errorf(op + " : hour must be between 0 and 23")
	}

	if timeString[2] != ':' {
		return nil, fmt.Errorf(op + " : time has invalid format")
	}

	minute1, err := strconv.Atoi(string(timeString[3]))
	if err != nil {
		return nil, fmt.Errorf(op + " : minute must be a number")
	}

	minute2, err := strconv.Atoi(string(timeString[4]))
	if err != nil {
		return nil, fmt.Errorf(op + " : minute must be a number")
	}
	minute := minute1*10 + minute2
	if minute < 0 || minute > 59 {
		return nil, fmt.Errorf(op + " : minute must be between 0 and 59")
	}

	return &Time{Hour: hour, Minute: minute}, nil
}

// mustParseEvent returns Event struct.
// if something goes wrong returns line
func mustParseEvent(eventLine string) *Event {
	event := strings.Split(eventLine, " ")
	if len(event) < 3 || len(event) > 4 {
		log.Fatal(eventLine)
	}

	eventTime, err := parseTime(event[0])
	if err != nil {
		log.Fatal(eventLine)
	}

	eventID, err := strconv.Atoi(event[1])
	if err != nil || eventID < 1 || eventID > 4 {
		log.Fatal(eventLine)
	}

	// we need to check: is nickname valid
	validSymbols := map[rune]bool{'0': true, '1': true, '2': true, '3': true, '4': true,
		'5': true, '6': true, '7': true, '8': true, '9': true, 'a': true, 'b': true, 'c': true, 'd': true,
		'e': true, 'f': true, 'g': true, 'h': true, 'i': true, 'j': true, 'k': true, 'l': true, 'm': true,
		'n': true, 'o': true, 'p': true, 'q': true, 'r': true, 's': true, 't': true, 'u': true, 'v': true,
		'w': true, 'x': true, 'y': true, 'z': true, '_': true, '-': true}
	clientName := event[2]
	for _, symbol := range clientName {
		if _, isValid := validSymbols[symbol]; !isValid {
			log.Fatal(eventLine)
		}
	}

	var eventTable int
	if len(event) == 4 {
		if eventID != 2 {
			log.Fatal(eventLine)
		}

		eventTable, err = strconv.Atoi(event[3])
		if err != nil || eventTable == 0 {
			log.Fatal(eventLine)
		}
	}

	return &Event{
		Time:   eventTime,
		ID:     eventID,
		Client: clientName,
		Table:  eventTable,
	}
}

type Time struct {
	Hour   int
	Minute int
}

// isFirstTimeEarlier compares two Time structure's
// return true if t1 < t2
// return false else
func isFirstTimeEarlier(t1, t2 *Time) bool {
	if t1.Hour < t2.Hour {
		return true
	} else if t1.Hour == t2.Hour && t1.Minute < t2.Minute {
		return true
	}
	return false
}

func makeStringFromTime(time *Time) string {
	if time.Hour < 10 && time.Minute < 10 {
		return fmt.Sprintf("0%d:0%d", time.Hour, time.Minute)
	}
	if time.Hour < 10 {
		return fmt.Sprintf("0%d:%d", time.Hour, time.Minute)
	}
	if time.Minute < 10 {
		return fmt.Sprintf("%d:0%d", time.Hour, time.Minute)
	}
	return fmt.Sprintf("%d:%d", time.Hour, time.Minute)
}

func addTwoTime(t1, t2 *Time) *Time {
	newMinute := t1.Minute + t2.Minute
	carry := newMinute / 60
	newMinute %= 60
	newHour := t1.Hour + t2.Hour + carry

	return &Time{
		Hour:   newHour,
		Minute: newMinute,
	}
}

type Queue struct {
	List           *list.List
	ClientsInQueue map[string]*list.Element
}

func New() *Queue {
	return &Queue{
		List:           list.New(),
		ClientsInQueue: make(map[string]*list.Element),
	}
}

func (q *Queue) PushBack(client string) {
	q.List.PushBack(client)
	q.ClientsInQueue[client] = q.List.Back()
}

func (q *Queue) Len() int {
	return q.List.Len()
}

func (q *Queue) GetFront() string {
	if q.List.Len() == 0 {
		return ""
	}
	cl := q.List.Front()
	q.List.Remove(cl)
	clientName := cl.Value.(string)
	delete(q.ClientsInQueue, clientName)
	return clientName
}

// remove element
func (q *Queue) Remove(client string) {
	cl := q.ClientsInQueue[client]
	delete(q.ClientsInQueue, client)
	q.List.Remove(cl)
}

func (q *Queue) IsInQueue(client string) bool {
	if _, ok := q.ClientsInQueue[client]; !ok {
		return false
	}
	return true
}

type Event1 struct {
	Event     *Event
	clubOpen  *Time
	clubClose *Time
	ClientMap *map[string]int
}

func playEvent1(e *Event1, writer *bufio.Writer) {
	timeString := makeStringFromTime(e.Event.Time)
	_, _ = writer.WriteString(timeString + " 1 " + e.Event.Client + "\n")

	// if client came before opening or after closing
	// returns error (NotOpenYet)
	if isFirstTimeEarlier(e.Event.Time, e.clubOpen) ||
		isFirstTimeEarlier(e.clubClose, e.Event.Time) {
		_, _ = writer.WriteString(timeString + " 13 " + ClubNotOpen + "\n")
		return
	}

	if _, InClub := (*e.ClientMap)[e.Event.Client]; InClub {
		_, _ = writer.WriteString(timeString + " 13 " + AlreadyInClub + "\n")
		return
	}

	(*e.ClientMap)[e.Event.Client] = 0
	return
}

type Event2 struct {
	Event     *Event
	Tables    *[]*Table
	Queue     *Queue
	ClientMap *map[string]int
}

func playEvent2(e *Event2, writer *bufio.Writer) {
	timeString := makeStringFromTime(e.Event.Time)
	tableString := strconv.Itoa(e.Event.Table)
	if e.Event.ID == 0 {
		_, _ = writer.WriteString(timeString + " 12 " + e.Event.Client + " " + tableString + "\n")
	} else {
		_, _ = writer.WriteString(timeString + " 2 " + e.Event.Client + " " + tableString + "\n")
	}

	// check for client existing
	if _, InClub := (*e.ClientMap)[e.Event.Client]; !InClub {
		_, _ = writer.WriteString(timeString + " 13 " + ClientNotExist + "\n")
		return
	}

	// table which client has chosen
	table := (*e.Tables)[e.Event.Table]

	if table.IsBusy {
		_, _ = writer.WriteString(timeString + " 13 " + PlaceIsBusy + "\n")
		return
	}

	// if client was sitting and tried to change
	// his place to free place
	if clientTable, _ := (*e.ClientMap)[e.Event.Client]; clientTable != NotSitting {

		// delete him from previous table
		delete(*e.ClientMap, e.Event.Client)

		// update fields of previous table
		prevTable := (*e.Tables)[clientTable]
		calculatePayment(prevTable, e.Event)
		// we must add subtract of sitting time and up
		periodOfPlaying := makeTimeFromMinutes(subtractTime(prevTable.LastSit, e.Event.Time))
		prevTable.WorkTime = addTwoTime(prevTable.WorkTime, periodOfPlaying)
		prevTable.IsBusy = false

		// if somebody in queue is waiting
		if e.Queue.Len() > 0 {

			clientFromQueue := e.Queue.GetFront()
			newEvent := &Event{
				Time:   e.Event.Time,
				ID:     0,
				Client: clientFromQueue,
				Table:  clientTable,
			}

			// play sitting event
			playEvent2(&Event2{
				Event:     newEvent,
				Tables:    e.Tables,
				ClientMap: e.ClientMap,
				Queue:     e.Queue,
			}, writer)
		}
	}

	table.IsBusy = true
	table.LastSit = e.Event.Time

	// now client sitting on new table
	(*e.ClientMap)[e.Event.Client] = e.Event.Table
}

func makeTimeFromMinutes(minutes int) *Time {
	hours := minutes / 60
	minutes %= 60
	return &Time{
		Hour:   hours,
		Minute: minutes,
	}
}

type Event3 struct {
	Event     *Event
	Tables    *[]*Table
	ClientMap *map[string]int
	Queue     *Queue
}

// В задании не было указано, что делать, если клиента нет в клубе
// и он решил ожидать. Поэтому я просто вывожу ClientUnknown.

// в задании ничего не сказано, что делать, если клиент ожидает,
// но уже находится в очереди - я ничего не делаю, он остаётся на том же месте
// я посчитал, что передвигать его в конец очереди - не логично

// в задании не сказано, что делать, если клиент ожидает, хотя есть свободные места
// но затем освобождается место - клиент на него сядет
// будем считать, что он ждал именно это место

// если клиент решил ожидать, сидя за столом - я считаю деньги за стол,
// клиента помещаю в конец очереди

func playEvent3(e *Event3, writer *bufio.Writer) {
	timeString := makeStringFromTime(e.Event.Time)
	_, _ = writer.WriteString(timeString + " 3 " + e.Event.Client + "\n")

	// check does client exist
	if _, InClub := (*e.ClientMap)[e.Event.Client]; !InClub {
		_, _ = writer.WriteString(timeString + " 13 " + ClientNotExist + "\n")
		return
	}

	// if client sitting on the table
	if place := (*e.ClientMap)[e.Event.Client]; place != NotSitting {
		// close table, sit somebody from queue if exist
		table := (*e.Tables)[place]

		table.IsBusy = false
		periodOfPlaying := makeTimeFromMinutes(subtractTime(table.LastSit, e.Event.Time))
		table.WorkTime = addTwoTime(table.WorkTime, periodOfPlaying)
		calculatePayment(table, e.Event)

		// clear client place
		(*e.ClientMap)[e.Event.Client] = NotSitting

		// if somebody in queue is waiting
		if e.Queue.Len() > 0 {

			clientFromQueue := e.Queue.GetFront()
			newEvent := &Event{
				Time:   e.Event.Time,
				ID:     0,
				Client: clientFromQueue,
				Table:  place,
			}

			// play sitting event
			playEvent2(&Event2{
				Event:     newEvent,
				Tables:    e.Tables,
				ClientMap: e.ClientMap,
			}, writer)
		}

		e.Queue.PushBack(e.Event.Client)
		return
	}

	for {
		for i, table := range *e.Tables {
			if table.IsBusy == false && i != 0 {
				_, _ = writer.WriteString(timeString + " 13 " + CantWait + "\n")
				return
			}
		}
		break
	}

	// if queue length bigger than number of tables
	if e.Queue.Len() >= len(*e.Tables)-1 {
		_, _ = writer.WriteString(timeString + " 11 " + e.Event.Client + "\n")
		// we need to delete client from base
		delete(*e.ClientMap, e.Event.Client)
		return
	}

	// add client in the queue

	// if client already in queue -> do nothing
	if e.Queue.IsInQueue(e.Event.Client) {
		return
	}

	e.Queue.PushBack(e.Event.Client)
}

type Event4 struct {
	Event     *Event
	Tables    *[]*Table
	ClientMap *map[string]int
	Queue     *Queue
}

func playEvent4(e *Event4, writer *bufio.Writer) {
	timeString := makeStringFromTime(e.Event.Time)
	_, _ = writer.WriteString(timeString + " 4 " + e.Event.Client + "\n")

	if _, InClub := (*e.ClientMap)[e.Event.Client]; InClub == false {
		_, _ = writer.WriteString(timeString + " 13 " + ClientNotExist + "\n")
		return
	}

	// client in the club
	// we need to check is he sitting somewhere
	if clientPlace := (*e.ClientMap)[e.Event.Client]; clientPlace == NotSitting {
		// check is he in queue
		if InQueue := e.Queue.IsInQueue(e.Event.Client); InQueue == true {
			// delete client from queue
			e.Queue.Remove(e.Event.Client)
		}
		delete(*e.ClientMap, e.Event.Client)
		return
	} else if e.Queue.Len() > 0 {

		// if client was sitting somewhere and queue not empty
		// we need to sit first people from queue to his place
		clientFromQueue := e.Queue.GetFront()

		// compute previous client
		table := (*e.Tables)[clientPlace]
		table.IsBusy = false
		periodOfPlaying := makeTimeFromMinutes(subtractTime(table.LastSit, e.Event.Time))
		table.WorkTime = addTwoTime(table.WorkTime, periodOfPlaying)
		calculatePayment(table, e.Event)

		// delete client
		delete(*e.ClientMap, e.Event.Client)

		// sit clientFromQueue
		newEvent := &Event{
			Time:   e.Event.Time,
			ID:     0,
			Client: clientFromQueue,
			Table:  clientPlace,
		}

		playEvent2(&Event2{
			Event:     newEvent,
			Tables:    e.Tables,
			ClientMap: e.ClientMap,
		}, writer)

		return

	} else {
		// if queue is empty
		table := (*e.Tables)[clientPlace]
		calculatePayment(table, e.Event)
		table.IsBusy = false
		periodOfPlaying := makeTimeFromMinutes(subtractTime(table.LastSit, e.Event.Time))
		table.WorkTime = addTwoTime(table.WorkTime, periodOfPlaying)

		// delete client
		delete(*e.ClientMap, e.Event.Client)
	}
}

// closeEvent just delete client from everywhere
// and print event 11
func closeEvent(e *Event4, writer *bufio.Writer) {
	timeString := makeStringFromTime(e.Event.Time)
	_, _ = writer.WriteString(timeString + " 11 " + e.Event.Client + "\n")

	// client in the club
	// if client in the queue
	if e.Queue.IsInQueue(e.Event.Client) {
		e.Queue.Remove(e.Event.Client)
		delete(*e.ClientMap, e.Event.Client)
		return
	}

	// client sitting
	// we need to compute his payment
	place := (*e.ClientMap)[e.Event.Client]
	if place != NotSitting {
		table := (*e.Tables)[place]

		calculatePayment(table, e.Event)
		periodOfPlaying := makeTimeFromMinutes(subtractTime(table.LastSit, e.Event.Time))
		table.WorkTime = addTwoTime(table.WorkTime, periodOfPlaying)
		table.IsBusy = false
	}

	// delete client
	delete(*e.ClientMap, e.Event.Client)
}

func calculatePayment(table *Table, e *Event) {
	sit, up := table.LastSit, e.Time
	minutesOnTheComputer := subtractTime(sit, up)

	// if client was sitting only 0 minute, -> must pay 1 time
	if minutesOnTheComputer == 0 {
		table.Payments++
		return
	}

	totalPayments := 0
	for minutesOnTheComputer > 0 {
		totalPayments++
		minutesOnTheComputer -= 60
	}

	table.Payments += totalPayments
}

// newTime = latter - earlier
// subtractTime returns minutes (int)
func subtractTime(earlier, latter *Time) int {
	time1 := latter.Hour*60 + latter.Minute
	time2 := earlier.Hour*60 + earlier.Minute
	return time1 - time2
}
