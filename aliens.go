package main

import (
    "log"
    "os"
    "bufio"
    "strings"
    "errors"
    "math/rand"
    "time"
    "flag"
    "fmt"
  )

// ---------------------  Types

//// --- Cordinal Point
type CardinalPoint int

const (
  north CardinalPoint = iota
  west
  south
  east
)

func (s CardinalPoint) String() string {
  return [...]string{"north", "west", "south", "east"}[s]
}

func parseCardinalPoint(s string) (cp CardinalPoint, err error) {
  err = nil
  switch s {
    case "north": cp = north
    case "west" : cp = west
    case "south" : cp = south
    case "east" : cp = east
    default     : err = errors.New("unrecognized cardinal point: " + s)
  }
  return cp, err
}

//// --- City Direction
type CityDirection struct {
  cardinalPoint CardinalPoint
  cityName string
}

func (d CityDirection) String() string {
  return d.cardinalPoint.String()+"="+d.cityName
}

func parseCityDirection(s string) (cd CityDirection, err error) {
  err = nil
  pd := strings.Split(s,"=")
  if len(pd) != 2 || len(pd[1]) == 0 {
    err = errors.New("unrecognized City direction: " + s)
    return
  }
  cp, err := parseCardinalPoint(pd[0])
  if err != nil {
    return
  }
  cd = CityDirection {cp, pd[1]}
  return
}

//// --- City
type City struct{
  name string
  directions []CityDirection
  alien *Alien
}

func (c City) String() (res string) {
  res = c.name
  for _, d := range c.directions {
    res += " " + d.String()
  }
  return
}

type Cities map[string]*City

func parseCitiesFromFile (fileName string) (cities Cities, err error) {
  var f *os.File
  f, err = os.OpenFile(fileName, os.O_RDONLY, 0755)
  if err != nil {
    log.Fatal(err)
    return
  }

  cities = make(Cities)
  scanner := bufio.NewScanner(f)
  for scanner.Scan() {
    flds := strings.Fields(scanner.Text())
    var ds []CityDirection
    for _, f := range flds[1:] {
      var cd CityDirection
      cd, err = parseCityDirection(f)
      if err != nil {
        log.Fatal("Parsing error: " + err.Error())
        return
      }
      ds = append (ds, cd)
    }
    cities[flds[0]] = &City {flds[0], ds, nil}
  }

  if err = scanner.Err(); err != nil {
    log.Fatal("Strange scanner error: " + err.Error())
    return
  }
  if err = f.Close(); err != nil {
    log.Fatal(err)
    return
  }
  return
}

//// --- Alien
type Alien struct {
  name int
  residence string
  isKilled bool
}

type Aliens []*Alien

func mkAliens (n int) Aliens {
  var aliens Aliens
  for i := 0; i<n; i++ {
    aliens = append(aliens, &Alien {i, "", false})
  }
  return aliens
}

// ---------------------  Logic

func destroyCity(cities Cities, city *City, alien *Alien) {
  alien.isKilled = true
  city.alien.isKilled = true
  delete(cities, city.name)
  fmt.Printf("%s has been destroyed by alien %d and alien %d!\n", city.name, city.alien.name, alien.name)
}

func move(cities Cities, alien *Alien) {
  if alien.residence == "" {
    alien.residence = getRandomCity(cities)
    curCity := cities[alien.residence]
    if curCity.alien != nil {
      destroyCity(cities, curCity, alien)
    } else {
      curCity.alien = alien
    }
  } else {
    if alien.isKilled {
      return
    }
    curCity := cities[alien.residence]
    var nextCity *City
    for {
      waysNum := len(curCity.directions)
      if waysNum == 0 {
        return
      }
      i := rand.Intn(waysNum)
      var ok bool
      nextCity, ok = cities[curCity.directions[i].cityName]
      if ok == true {
        break;
      }
      curCity.directions = append(curCity.directions[:i], curCity.directions[i+1:]...)
    }

    alien.residence = nextCity.name
    if nextCity.alien == nil || nextCity.alien.isKilled {
      nextCity.alien = alien
      curCity.alien = nil
    } else {
      destroyCity(cities, nextCity, alien)
      curCity.alien = nil
    }
  }
}

// ---------------------  Helpers

func getRandomCity (cities Cities) string {
  cityNames := make([]string, 0, len(cities))
  for k := range cities {
      cityNames = append(cityNames, k)
  }
  i := rand.Intn(len(cities))
  return cityNames[i]
}

func filterKilled (aliens Aliens) (Aliens){
  n := 0
  for _, x := range aliens {
    if !x.isKilled {
      aliens[n] = x
      n++
    }
  }
  aliens = aliens[:n]
  return aliens
}

func filterDestroyedCities (cities Cities) {
  for _, c := range cities {
    n := 0
    for _, d := range c.directions {
      _, exists := cities[d.cityName]
      if exists {
        c.directions[n] = d
        n++
      }
    }
    c.directions = c.directions[:n]
  }
}

func parseFlags () (aliensNumber int, fileName string) {
  flag.IntVar(&aliensNumber, "N", 3, "Number of aliens" )
  flag.StringVar(&fileName, "file", "cities.txt", "File with map")
  flag.Parse()
  return
}

// ---------------------  Main

func main() {
  rand.Seed(time.Now().UnixNano())

  aliensNumber, mapFile := parseFlags()
  cities, err := parseCitiesFromFile(mapFile)
  if err != nil {
    return
  }

  aliens := mkAliens(aliensNumber)

  for i := 0; i < 10000; i++ {
    for _, alien := range aliens {
      move(cities, alien)
    }
    aliens = filterKilled(aliens)
    if len(aliens) == 0 {
      break;
    }
  }

  filterDestroyedCities(cities)
  for _,c := range cities {
    fmt.Printf("%v\n",c)
  }
}
