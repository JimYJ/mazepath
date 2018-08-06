package maps

import ()
import "runtime"
import "log"

/*
 *使用y坐标轴反向的坐标轴
 */

var (
	startX, startY   = 0, 0 //初始坐标轴
	endX, endY, w, h int
	stepQueue        = make(chan *step, 100)
	exploredCoord    = make(map[coord]int)
	finish           = make(chan bool, 0)
	wait             = make(chan bool, 0)
)

type step struct {
	cod   *coord   //当前横轴,竖轴坐标
	path  []*coord //路径
	steps int      //当前第几步
}

type coord struct {
	x, y int //当前横轴,竖轴坐标
}

// GetPath 计算地图路径
func GetPath() {
	w, h := getMapWH(mazeMap)
	endX, endY = w-1, h-1
	go stepPath(mazeMap)
	start := &coord{0, 0}
	exploredCoord[*start] = 0
	stepQueue <- &step{
		start,
		[]*coord{start},
		0,
	}
	<-wait
}

// 获得地图高宽
func getMapWH(maze [][]int) (int, int) {
	h = len(maze)
	w = len(maze[0])
	log.Println(w, h)
	return w, h
}

// 获得当前位置的四个方向坐标,横轴x,竖轴y
func getNextStep(curStep *step) (*coord, *coord, *coord, *coord) {
	if curStep != nil {
		var up, down, left, right = &coord{}, &coord{}, &coord{}, &coord{}
		up.x, up.y = curStep.cod.x, curStep.cod.y-1
		down.x, down.y = curStep.cod.x, curStep.cod.y+1
		left.x, left.y = curStep.cod.x-1, curStep.cod.y
		right.x, right.y = curStep.cod.x+1, curStep.cod.y
		return up, down, left, right
	}
	return nil, nil, nil, nil
}

// 根据坐标轴探测四个方向的路径是否可用
func stepPath(maze [][]int) {
	for {
		select {
		case curStep := <-stepQueue:
			up, down, left, right := getNextStep(curStep)
			checkCoord(up, curStep, maze)
			checkCoord(down, curStep, maze)
			checkCoord(left, curStep, maze)
			checkCoord(right, curStep, maze)
		case <-finish:
			log.Println(444)
			log.Println(runtime.NumGoroutine())
			break
		}
	}
}

// 检查坐标
func checkCoord(cod *coord, curStep *step, maze [][]int) {
	if checkRange(cod) {
		if checkNewStep(cod, maze) {
			if _, ok := exploredCoord[*cod]; ok {
				return
			}
			exploredCoord[*cod] = curStep.steps + 1
			newStep := &step{}
			newStep.cod = cod
			newStep.path = append(curStep.path, cod)
			newStep.steps = curStep.steps + 1
			if cod.x == endX && cod.y == endY {
				log.Println(222)
				go finishQueue(newStep)
			} else {
				stepQueue <- newStep
				log.Println(cod)
			}
		}
	}
}

// 检查索引是否越界
func checkRange(cod *coord) bool {
	if cod != nil {
		if cod.x >= 0 && cod.x < w && cod.y >= 0 && cod.y < h {
			return true
		}
		return false
	}
	return false
}

// 检查坐标是否为可通过路径
func checkNewStep(cod *coord, maze [][]int) bool {
	if maze[cod.y][cod.x] == 0 {
		return true
	}
	return false
}

func finishQueue(newStep *step) {
	close(stepQueue)
	for {
		_, isClose := <-stepQueue
		if !isClose {
			break
		}
	}
	finish <- true
	for _, v := range newStep.path {
		log.Println(v)
	}
	wait <- true
}
