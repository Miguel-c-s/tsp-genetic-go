package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"
)

// NodeSize is the size of the node circles in pixels
const NodeSize = 10

// City represents a city with x and y coordinates
type City struct {
	x int
	y int
}

// Distance returns the distance to another city
func (city *City) Distance(other *City) float64 {
	xDistance := math.Abs(float64(city.x - other.x))
	yDistance := math.Abs(float64(city.y - other.y))
	return math.Sqrt(xDistance*xDistance + yDistance*yDistance)
}

// Tour represents a tour of cities
type Tour struct {
	cities   []*City
	distance float64
	fitness  float64
}

// NewTour creates a new tour with a random permutation of cities
func NewTour(cities []*City) *Tour {
	tour := &Tour{
		cities: make([]*City, len(cities)),
	}
	copy(tour.cities, cities)
	rand.Shuffle(len(tour.cities), func(i, j int) {
		tour.cities[i], tour.cities[j] = tour.cities[j], tour.cities[i]
	})
	tour.distance = tour.CalculateDistance()
	return tour
}

// CalculateDistance calculates the total distance of the tour
func (tour *Tour) CalculateDistance() float64 {
	distance := 0.0
	for i := 0; i < len(tour.cities)-1; i++ {
		distance += tour.cities[i].Distance(tour.cities[i+1])
	}
	distance += tour.cities[len(tour.cities)-1].Distance(tour.cities[0])
	return distance
}

// Mutate swaps two random cities in the tour
func (tour *Tour) Mutate() {
	i := rand.Intn(len(tour.cities))
	j := rand.Intn(len(tour.cities))
	tour.cities[i], tour.cities[j] = tour.cities[j], tour.cities[i]
	tour.distance = tour.CalculateDistance()
}

// contains returns true if the given city is in the given slice of cities
func contains(cities []*City, city *City) bool {
	for _, c := range cities {
		if c == city {
			return true
		}
	}
	return false
}

// Crossover creates a new tour by crossing over two tours at a random point
func Crossover(tour1, tour2 *Tour) *Tour {
	newTour := &Tour{
		cities: make([]*City, len(tour1.cities)),
	}
	for i := 0; i < len(newTour.cities)/2; i++ {
		newTour.cities[i] = tour1.cities[i]
	}
	i := len(newTour.cities) / 2
	j := 0
	for j < len(tour2.cities) {
		if !contains(newTour.cities, tour2.cities[j]) {
			newTour.cities[i] = tour2.cities[j]
			i++
		}
		j++
	}
	newTour.distance = newTour.CalculateDistance()
	return newTour
}

// Evolve runs the genetic algorithm on the given population of tours
func Evolve(population []*Tour, crossoverRate float64, mutationRate float64, cities []*City) []*Tour {
	newPopulation := make([]*Tour, len(population))

	for i := 0; i < len(newPopulation); i++ {
		if i < len(population)/2 {
			newPopulation[i] = population[i]
		} else {
			tour1 := SelectTour(population)
			tour2 := SelectTour(population)
			if rand.Float64() < crossoverRate {
				newPopulation[i] = Crossover(tour1, tour2)
			} else {
				newPopulation[i] = NewTour(cities)
			}
			if rand.Float64() < mutationRate {
				newPopulation[i].Mutate()
			}
		}
	}
	sort.Slice(newPopulation, func(i, j int) bool {
		return newPopulation[i].distance < newPopulation[j].distance
	})

	return newPopulation
}

// SelectTour selects a tour from the given population using roulette wheel selection
func SelectTour(population []*Tour) *Tour {
	fitnessSum := 0.0
	for _, tour := range population {
		fitnessSum += tour.fitness
	}
	randNum := rand.Float64() * fitnessSum
	curSum := 0.0
	for _, tour := range population {
		curSum += tour.fitness
		if curSum >= randNum {
			return tour
		}
	}
	return population[0]
}

// DrawTour draws the given tour on the image
func DrawTour(tour *Tour, img *image.RGBA) {
	for _, city := range tour.cities {
		for x := city.x - NodeSize; x <= city.x+NodeSize; x++ {
			for y := city.y - NodeSize; y <= city.y+NodeSize; y++ {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}

	for i := 0; i < len(tour.cities)-1; i++ {
		DrawLine(tour.cities[i], tour.cities[i+1], img)
	}
	DrawLine(tour.cities[len(tour.cities)-1], tour.cities[0], img)
}

// DrawLine draws a line between the two cities on the image
func DrawLine(city1, city2 *City, img *image.RGBA) {
	dx := city2.x - city1.x
	dy := city2.y - city1.y
	if dx == 0 {
		if dy > 0 {
			for y := city1.y; y <= city2.y; y++ {
				img.Set(city1.x, y, color.RGBA{0, 0, 0, 255})
			}
		} else {
			for y := city2.y; y <= city1.y; y++ {
				img.Set(city1.x, y, color.RGBA{0, 0, 0, 255})
			}
		}
		return
	}
	slope := float64(dy) / float64(dx)
	if math.Abs(slope) > 1 {
		if dy > 0 {
			for y := city1.y; y <= city2.y; y++ {
				x := int(float64(y-city1.y)/slope + float64(city1.x))
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		} else {
			for y := city2.y; y <= city1.y; y++ {
				x := int(float64(y-city1.y)/slope + float64(city1.x))
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	} else {
		if dx > 0 {
			for x := city1.x; x <= city2.x; x++ {
				y := int(slope*float64(x-city1.x) + float64(city1.y))
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		} else {
			for x := city2.x; x <= city1.x; x++ {
				y := int(slope*float64(x-city1.x) + float64(city1.y))
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
}

// CalculateFitness calculates the fitness of a tour based on its distance
func (tour *Tour) CalculateFitness() {
	tour.fitness = 1.0 / tour.distance
}

func main() {
	const numProblems = 6
	const numCities = 32
	const maxX, maxY = 256, 256
	const numThreads = 12

	// Generate random cities
	var problems [numProblems][]*City
	for i := 0; i < numProblems; i++ {
		problems[i] = make([]*City, numCities)
		for j := 0; j < numCities; j++ {
			problems[i][j] = &City{
				x: rand.Intn(maxX),
				y: rand.Intn(maxY),
			}
		}
	}

	// Open file to write results to
	file, err := os.Create("results.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a new CSV file
	csvFile, err := os.Create("results.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	// Create a new CSV writer
	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	// Write the headers
	headers := []string{"Thread Name", "Num Generations", "Population Size", "Mutation Rate", "Crossover Rate", "Problem Num", "Distance", "Elapsed Time"}
	if err := csvWriter.Write(headers); err != nil {
		log.Fatal(err)
	}

	// Create channels to hold data for each problem
	type result struct {
		threadName     string
		numGenerations int
		populationSize int
		mutationRate   float64
		crossoverRate  float64
		problemNum     int
		distance       float64
		elapsedTime    time.Duration
	}
	results := make(chan result, numProblems)

	// Create and start threads
	for i := 0; i < numThreads; i++ {
		go func(threadName string) {
			for problemNum, cities := range problems {
				numGenerations := 100000
				populationSize := 100
				mutationRate := 0.05  // rand.Float64()
				crossoverRate := 0.70 // rand.Float64()

				start := time.Now()
				population := make([]*Tour, populationSize)
				for i := 0; i < populationSize; i++ {
					population[i] = NewTour(cities)
				}

				for i := range population {
					population[i].CalculateFitness()
				}
				fitnessSum := 0.0
				for i := range population {
					fitnessSum += population[i].fitness
				}
				for i := range population {
					population[i].fitness /= fitnessSum
				}

				for i := 0; i < numGenerations; i++ {
					for _, tour := range population {
						tour.fitness = 1.0 / tour.distance
					}

					population = Evolve(population, crossoverRate, mutationRate, cities)
				}

				bestTour := population[0]
				for _, tour := range population {
					if tour.distance < bestTour.distance {
						bestTour = tour
					}
				}
				img := image.NewRGBA(image.Rect(0, 0, 256, 256))
				DrawTour(bestTour, img)
				f, _ := os.Create(fmt.Sprintf("./images/tour_thread_%s_problem_%d.png", threadName, problemNum))
				defer f.Close()
				png.Encode(f, img)

				fmt.Println("Thread:", threadName, "Problem:", problemNum, "Distance:", bestTour.distance, "Time:", time.Since(start))
				// Send result to channel
				results <- result{
					threadName:     threadName,
					numGenerations: numGenerations,
					populationSize: populationSize,
					mutationRate:   mutationRate,
					crossoverRate:  crossoverRate,
					problemNum:     problemNum,
					distance:       bestTour.distance,
					elapsedTime:    time.Since(start),
				}
			}
		}(fmt.Sprintf("Thread-%d", i+1))
	}

	// Create a map to store the results
	resultsMap := make(map[int][]result)

	// Write the results to the map
	for i := 0; i < numProblems*numThreads; i++ {
		result := <-results
		resultsMap[result.problemNum] = append(resultsMap[result.problemNum], result)
	}

	// Write results to file
	for i := 0; i < numProblems; i++ {
		for j := 0; j < len(resultsMap[i]); j++ {
			result := resultsMap[i][j]
			_, err = fmt.Fprintf(file, "Thread Name: %s\nGenetic Parameters: numGenerations=%d populationSize=%d mutationRate=%f crossoverRate=%f\nProblem #%d\nDistance: %f\nTime taken: %s\n\n", result.threadName, result.numGenerations, result.populationSize, result.mutationRate, result.crossoverRate, result.problemNum, result.distance, result.elapsedTime)
			if err != nil {
				log.Fatal(err)
			}

			record := []string{
				result.threadName,
				strconv.Itoa(result.numGenerations),
				strconv.Itoa(result.populationSize),
				strconv.FormatFloat(result.mutationRate, 'f', 6, 64),
				strconv.FormatFloat(result.crossoverRate, 'f', 6, 64),
				strconv.Itoa(result.problemNum),
				strconv.FormatFloat(result.distance, 'f', 6, 64),
				result.elapsedTime.String(),
			}
			if err := csvWriter.Write(record); err != nil {
				log.Fatal(err)
			}
		}
	}
}
