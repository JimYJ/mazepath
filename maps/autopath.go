package maps

import ()
import "time"
import "fmt"
import "runtime"
import "log"

/*
 *使用y坐标轴反向的坐标轴
 */

var (
	startX, startY   = 0, 0 //初始坐标轴
	endX, endY, w, h int
	stepQueue        = make(chan *step, 1000)
	exploredCoord    = make(map[coord]int)
	finish           = make(chan bool, 0)
	wait             = make(chan bool, 0)
	pathCount        []int //路径分支线记号统计
	end              bool
)

type step struct {
	cod   *coord        //当前横轴,竖轴坐标
	path  map[coord]int //路径
	steps int           //当前第几步
	no    int           //当前路线分支线的编号
}

type coord struct {
	x, y int //当前横轴,竖轴坐标
}

// GetPath 计算地图路径
func GetPath() {
	t := time.Now()
	w, h := getMapWH(mazeMap)
	endX, endY = w-1, h-1
	go stepPath(mazeMap)
	start := &coord{0, 0}
	exploredCoord[*start] = 0
	startPath := make(map[coord]int)
	startPath[*start] = 0
	pathCount = []int{1}
	stepQueue <- &step{
		start,
		startPath,
		0,
		1,
	}
	<-wait
	log.Println("总共耗时:", time.Now().Sub(t))
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
	i := 1
	for {
		select {
		case curStep := <-stepQueue:
			up, down, left, right := getNextStep(curStep)
			var lastCod *coord
			var lastStep int
			lastPath := make(map[coord]int)
			if up != nil && down != nil && left != nil && right != nil && curStep != nil {
				lastCod = curStep.cod
				lastPath = copyMap(curStep.path)
				lastStep = curStep.steps
				log.Println(fmt.Sprintf("开始第%d步探索,当前分支路径编号:%d，当前坐标为:%v,探索的四个坐标为:%v,%v,%v,%v", i, curStep.no, *lastCod, *up, *down, *left, *right))
			}
			rs1 := checkCoord(up, lastCod, curStep, maze, lastPath, lastStep)
			rs2 := checkCoord(left, lastCod, curStep, maze, lastPath, lastStep)
			rs3 := checkCoord(down, lastCod, curStep, maze, lastPath, lastStep)
			rs4 := checkCoord(right, lastCod, curStep, maze, lastPath, lastStep)
			i++
			if rs1 == false && rs2 == false && rs3 == false && rs4 == false {
				if curStep != nil {
					log.Println(fmt.Sprintf("第%d条分支无路可走，销毁！最后坐标:%v", curStep.no, lastCod))
					// printMap(curStep.path)
					curStep = nil
				}
			}
		case <-finish:
			log.Println(runtime.NumGoroutine())
			break
		}
	}
}

// 检查坐标
func checkCoord(cod, lastCod *coord, curStep *step, maze [][]int, lastPath map[coord]int, lastStep int) bool {
	if checkRange(cod) {
		if checkNewStep(cod, maze) {
			if _, ok := curStep.path[*cod]; ok {
				return false
			}
			if v, ok := exploredCoord[*cod]; ok {
				if lastStep+1 >= v {
					return false
				}
			}
			exploredCoord[*cod] = lastStep + 1
			var s *step
			if (*curStep.cod) != (*lastCod) {
				newStep := &step{}
				newStep.cod = cod
				newStep.path = lastPath
				newStep.path[*cod] = lastStep + 1
				newStep.steps = lastStep + 1
				newStep.no = len(pathCount) + 1
				log.Println(fmt.Sprintf("遇到分叉，生成新的路径分支，编号:%d,当前坐标%v，上一步坐标:%v", len(pathCount)+1, cod, lastCod))
				pathCount = append(pathCount, len(pathCount)+1)
				s = newStep
			} else {
				curStep.cod = cod
				curStep.path[*cod] = curStep.steps + 1
				curStep.steps++
				s = curStep
			}
			if cod.x == endX && cod.y == endY {
				end = true
				go finishQueue(s)
				return true
			} else {
				if !end {
					stepQueue <- s
				}
				return true
			}
		}
		return false
	}
	return false
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
	log.Println("找出最短路径的分支编号是:", newStep.no)
	printMap(newStep.path)
	wait <- true
}

func checkPathRepeat(cod *coord, path []*coord) bool {
	rs := false
	for i := 0; i < len(path); i++ {
		if (*path[i]) == (*cod) {
			rs = true
			break
		}
	}
	return rs
}

func printMap(m map[coord]int) {
	if m != nil {
		log.Println(pathCount)
		for i := 0; i < len(m); i++ {
			for k, v := range m {
				if v == i {
					log.Println(fmt.Sprintf("第%d步，坐标%v", v, k))
					// continue
				}
			}
		}
	}
}

func copyMap(m map[coord]int) map[coord]int {
	n := make(map[coord]int)
	for k, v := range m {
		n[k] = v
	}
	return n
}
