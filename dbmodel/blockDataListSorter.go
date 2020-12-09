package dbmodel

// 排序key
type BlockDataListSorterSortKey int

const (
	BlockDataListSorterSortKeySortByHeight = BlockDataListSorterSortKey(iota) // 根据高度排序
)

// 区块列表排序器
type BlockDataListSorter struct {
	list    []*BlockData               // 区块数据
	sortKey BlockDataListSorterSortKey // 排序key
}

// NewBlockDataListSorter 工厂方法
func NewBlockDataListSorter(list []*BlockData) BlockDataListSorter {
	return BlockDataListSorter{
		list:    list,
		sortKey: BlockDataListSorterSortKeySortByHeight,
	}
}

// Len 长度
func (object BlockDataListSorter) Len() int {
	return len(object.list)
}

// Swap交换
func (object BlockDataListSorter) Swap(i, j int) {
	object.list[i], object.list[j] = object.list[j], object.list[i]
}

// Less 小于
func (object BlockDataListSorter) Less(i, j int) bool {
	return object.list[i].Height < object.list[j].Height
}
