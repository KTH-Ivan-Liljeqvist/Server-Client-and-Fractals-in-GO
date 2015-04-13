/*
* author - Stefan Nilsson
* Ivan Liljeqvist answered the questions at the end of this file.
* The questions are in Swedish.
*
 */

// http://www.nada.kth.se/~snilsson/concurrency/

package main

import (
	"fmt"
	"sync"
)

// This programs demonstrates how a channel can be used for sending and
// receiving by any number of goroutines. It also shows how  the select
// statement can be used to choose one out of several communications.
func main() {
	people := []string{"Anna", "Bob", "Cody", "Dave", "Eva"}
	match := make(chan string, 1) // Make room for one unmatched send.

	wg := new(sync.WaitGroup)
	wg.Add(len(people))
	for _, name := range people {
		go Seek(name, match, wg)
	}
	wg.Wait()
	select {
	case name := <-match:
		fmt.Printf("No one received %s's message.\n", name)
	default:
		// There was no pending send operation.
	}
}

// Seek either sends or receives, whichever possible, a name on the match
// channel and notifies the wait group when done.
func Seek(name string, match chan string, wg *sync.WaitGroup) {
	select {
	case peer := <-match:
		fmt.Printf("%s sent a message to %s.\n", peer, name)
	case match <- name:

		// Wait for someone to receive my message.
	}
	wg.Done()
}

/*

** Vad händer om man tar bort go-kommandot från Seek-anropet i main-funktionen?

Hypotes:
När man avänder nyckelordet go startar man funktionen som en ny tråd.
Tar man bort det kommer anropen på Seek att komma i ordning efter varandra.
På det sättet kommer Anna alltid skicka meddelande till Bob och Cody till Dave medan ingen kommer ta emot Evas.
Skulle man ha med 'go' kan ordningen slumpas och kan inte vara säkra vem som kommer skicka och ta emot vad.

Experiment:
Jag tog bort 'go' och fick samma output som med 'go'.
Just i mitt fall fick jag samma ordning i anropen som med go. Men teoretiskt är denna ordning inte garanterad.

** Vad händer om man byter deklarationen wg := new(sync.WaitGroup) mot var wg sync.WaitGroup och parametern wg *sync.WaitGroup mot wg sync.WaitGroup?

Hypotes:
Man andrärar om en pekare till en vanlig variabel. När man sedan ger variabeln wg till funktionen Seek kommer originalet inte att följa med - det blir en kopia.
På det sättet kommer vi aldrig komma förbi wg.Wait() i main metoden eftersom done() körs aldrig på original versionen av wg.

Experiment:
Det hela resulterade i en deadlock.
Vi kommer inte vidare förbi wg.Wait() vilket blockerar main metoden och alla Seek rutiner pausas.


** Vad händer om man tar bort bufferten på kanalen match?

Hypotes:
När vi har en buffrad kanal innebär det att programmet kan fortsätta att köras även när det ligger något i kanalen och man inte tar ut det.
I vårt fall kommer vi sätta in Eva och ingen kommer att ta ut hennes 'meddelande'.
Programmet kommer frysas eftersom den obuffrade kanalen kommer att blockera exekveringen och det hela kommer resultera i deadlock.

Experiment:
När jag gjorde kanalen obuffrad resulterade det hela i deadlock. Jag testade ta bort Eva ur array:en och då fick man samma output som innan.
Alltså är det Evas meddelande som inte tas ut ur kanalen som blockerar programmet.

** Vad händer om man tar bort default-fallet från case-satsen i main-funktionen?

Hypotes:
Select blockerar programmet förrän det kan köra något av case-scenarion.
Om man tar bort default kommer det bara finnas ett alternativ - att läsa från kanalen.
Om vi inte har något att läsa kommer programmet blockeras eftersom vi inte kommer komma förbi select.
I vårat fall ligger Evas meddelande kvar i kanalen och därför tror jag programmet kommer att köras som vanligt.
Om man har ett jämnt antal namn i arrayen kommer det bli deadlock eftersom programmet kommer frysas för att det inte finns något att läsa
från kanalen.

Experiment:
Med Eva körs programmet som vanligt. Utan Eva kraschar det. Lägger man in en ny person efter Eva kraschar det också.
Alltså stämmer min hypotes att när det är jämnt antal personer kan dem läsa varandras meddelande och då blir kanalen tom
och vi kan inte komma vidare i programmet eftersom det inte finns något att ta ut ur den. Med default-fallet kan vi komma vidare även om
kanalen är tom.

*/
