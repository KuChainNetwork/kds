package dbmodel

// 排序key
type BlockListSorterSortKey int

const (
	BlockListSorterSortKeySortByHeight = BlockListSorterSortKey(iota) // 根据高度排序
)

// 区块列表排序器
type BlockListSorter struct {
	list    []*Block               // 区块数据
	sortKey BlockListSorterSortKey // 排序key
}

// NewBlockListSorter 工厂方法
func NewBlockListSorter(list []*Block) BlockListSorter {
	return BlockListSorter{
		list:    list,
		sortKey: BlockListSorterSortKeySortByHeight,
	}
}

// Len 长度
func (object BlockListSorter) Len() int {
	return len(object.list)
}

// Swap交换
func (object BlockListSorter) Swap(i, j int) {
	object.list[i], object.list[j] = object.list[j], object.list[i]
}

// Less 小于
func (object BlockListSorter) Less(i, j int) bool {
	return object.list[i].Height < object.list[j].Height
}
