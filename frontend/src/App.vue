<template>
  <el-config-provider :namespace="namespace">
    <div class="app-root">
      <div class="container">
        <el-container>
          <el-header>
            <h2 class="app-title">Smart Clipboard</h2>
            <el-button type="text" @click="showSettings">
              <el-icon size="large"><Setting /></el-icon>
            </el-button>
          </el-header>
          
          <!-- 添加标签栏 -->
          <div class="tags-bar">
            <div class="tags-list">
              <div 
                class="tag-item all-tag"
                :class="{ active: currentTag === null }"
                @click="currentTag = null"
              >
                全部
              </div>
              <div
                v-for="tag in tags"
                :key="tag.id"
                class="tag-item"
                :class="{ active: currentTag === tag.id, dragging: draggingTag === tag.id }"
                @click="currentTag = tag.id"
                @contextmenu.prevent="showTagContextMenu($event, tag)"
                @dragover.prevent
                @drop.stop="handleDrop($event, tag.id)"
                draggable="true"
                @dragstart="handleTagDragStart($event, tag)"
                @dragend="handleTagDragEnd"
                @dragenter.prevent="handleTagDragEnter(tag)"
              >
                <div class="tag-dot" :style="{ backgroundColor: tag.color }"></div>
                {{ tag.name }}
              </div>
            </div>
            <!-- 添加搜索部分 -->
            <div class="search-container">
              <el-button 
                v-if="!showSearch"
                type="text" 
                class="search-btn"
                @click="toggleSearch"
              >
                <el-icon><Search /></el-icon>
              </el-button>
              <div v-else class="search-input-container">
                <el-input
                  v-model="searchText"
                  placeholder="搜索内容"
                  class="search-input"
                  ref="searchInput"
                  @keydown.esc="handleSearchEsc"
                  @keydown.enter="handleSearchEnter"
                  @input="handleSearchInput"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                  <template #suffix>
                    <el-icon 
                      class="clear-search" 
                      @click="clearSearch"
                      v-if="searchText"
                    >
                      <Close />
                    </el-icon>
                  </template>
                </el-input>
              </div>
            </div>
            <el-button 
              type="text" 
              class="add-tag-btn"
              @click="showAddTagDialog"
            >
              <el-icon><Plus /></el-icon>
            </el-button>
          </div>

          <el-main>
            <div class="clipboard-items" ref="itemsContainer" tabindex="0" @keydown="handleKeyDown">
              <template v-if="filteredHistory.length > 0">
                <div
                  v-for="(item, index) in filteredHistory"
                  :key="item.id"
                  :class="['clipboard-item', { 'selected': index === selectedIndex }]"
                  @click="selectItem(index)"
                  @dblclick="copyContent(item)"
                  draggable="true"
                  @dragstart="handleDragStart($event, item)"
                >
                  <!-- 添加标签指示器 -->
                  <div 
                    v-if="item.tagId" 
                    class="item-tag-indicator"
                    :style="{ backgroundColor: getTagColor(item.tagId) }"
                  ></div>
                  
                  <!-- 现有的内容展示部分 -->
                  <template v-if="item.type === 'text'">
                    <div class="text-content">{{ truncateText(item.content, 200) }}</div>
                  </template>
                  <template v-else-if="item.type === 'image'">
                    <img :src="item.content" class="clipboard-image" />
                  </template>
                  <div class="item-timestamp">{{ formatDate(item.timestamp) }}</div>
                </div>
              </template>
              <div v-else class="empty-state">
                <el-icon class="empty-icon"><DocumentCopy /></el-icon>
                <p class="empty-text">{{ getEmptyStateText }}</p>
              </div>
            </div>
          </el-main>
        </el-container>

        <!-- 修改添加标签对话框部分 -->
        <el-dialog v-model="addTagVisible" title="新增标签" width="30%">
          <el-form>
            <el-form-item>
              <div class="tag-input-group">
                <el-color-picker 
                  v-model="newTagColor" 
                  size="small"
                  class="tag-color-picker"
                />
                <el-input 
                  v-model="newTagName" 
                  placeholder="请输入标签名称"
                  class="tag-name-input"
                />
              </div>
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="addTagVisible = false">取消</el-button>
              <el-button type="primary" @click="addTag">确定</el-button>
            </span>
          </template>
        </el-dialog>

        <!-- 现有的设置对话框 -->
        <el-dialog v-model="settingsVisible" title="设置" width="30%">
          <el-form>
            <el-form-item label="最大历史记录数">
              <el-input-number 
                v-model="maxHistory" 
                :min="1" 
                :max="50"
                class="max-history-input" 
              />
            </el-form-item>
            <el-form-item label="复制后自动隐藏">
              <el-switch v-model="autoHide" />
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="settingsVisible = false">取消</el-button>
              <el-button type="primary" @click="saveSettings">确定</el-button>
            </span>
          </template>
        </el-dialog>

        <!-- 添加标签编辑对话框 -->
        <el-dialog v-model="editTagVisible" title="编辑标签" width="30%">
          <el-form>
            <el-form-item>
              <div class="tag-input-group">
                <el-color-picker 
                  v-model="editingTag.color" 
                  size="small"
                  class="tag-color-picker"
                />
                <el-input 
                  v-model="editingTag.name" 
                  placeholder="请输入标签名称"
                  class="tag-name-input"
                />
              </div>
            </el-form-item>
          </el-form>
          <template #footer>
            <span class="dialog-footer">
              <el-button @click="editTagVisible = false">取消</el-button>
              <el-button type="primary" @click="saveTagEdit">确定</el-button>
            </span>
          </template>
        </el-dialog>

        <!-- 添加悬浮菜单 -->
        <div 
          v-if="showContextMenu" 
          class="tag-popup-menu"
          :style="contextMenuStyle"
          @mouseleave="hideContextMenu"
        >
          <div class="menu-item" @click="editTag">
            <el-icon><Edit /></el-icon>
            编辑
          </div>
          <div class="menu-item delete" @click="deleteTag">
            <el-icon><Delete /></el-icon>
            删除
          </div>
        </div>
      </div>
    </div>
  </el-config-provider>
</template>

<script>
import { Setting, WarningFilled, Plus, Close, Search, Edit, Delete, DocumentCopy } from '@element-plus/icons-vue'
import { ref, onMounted, watch } from 'vue'
import { useDark } from '@vueuse/core'

// 检测系统主题
const isDark = useDark()

export default {
  components: {
    Setting,
    WarningFilled,
    Plus,
    Close,
    Search,
    Edit,
    Delete,
    DocumentCopy
  },
  data() {
    return {
      history: [],
      settingsVisible: false,
      maxHistory: 50,
      selectedIndex: -1,
      currentTag: null,
      addTagVisible: false,
      newTagName: '',
      newTagColor: '#409EFF',
      tags: [],
      showSearch: false,
      searchText: '',
      editTagVisible: false,
      editingTag: {
        id: '',
        name: '',
        color: ''
      },
      showContextMenu: false,
      contextMenuPosition: { x: 0, y: 0 },
      activeTag: null,
      draggingTag: null,
      dragOverTag: null,
      config: {
        maxHistory: 50,
        tags: []
      },
      currentTagIndex: -1,
      autoHide: false,
    }
  },
  computed: {
    filteredHistory() {
      let result = this.currentTag ? 
        this.history.filter(item => item.tagId === this.currentTag) :
        this.history

      if (this.searchText) {
        const searchLower = this.searchText.toLowerCase().trim()
        result = result.filter(item => {
          // 只对文本类型进行搜索
          if (item.type === 'text') {
            return item.content.toLowerCase().includes(searchLower)
          }
          // 图片类型不参与搜索过滤
          return false
        })
      }

      return result
    },
    contextMenuStyle() {
      return {
        left: this.contextMenuPosition.x + 'px',
        top: this.contextMenuPosition.y + 'px'
      }
    },
    getEmptyStateText() {
      if (this.searchText) {
        return '没有找到匹配的内容'
      }
      if (this.currentTag) {
        return '该标签下暂无内容'
      }
      return '暂无复制记录'
    },
    allTags() {
      return [{ id: null, name: '全部' }, ...this.tags]
    }
  },
  async created() {
    await this.loadHistory()
    await this.loadConfig()
    setInterval(this.loadHistory, 3000)
    
    window.runtime.EventsOn("toggleWindow", () => {
      window.go.main.App.ToggleWindow()
      // 当窗口显示时，检查并选中第一个卡片
      this.$nextTick(() => {
        if (this.filteredHistory.length > 0 && this.selectedIndex === -1) {
          this.selectedIndex = 0
        }
      })
    })
    
    // 添加历史记录更新事件监听
    window.runtime.EventsOn("historyUpdated", () => {
      this.loadHistory()
    })

    // 添加错误提示监听
    window.runtime.EventsOn("clipboardError", (message) => {
      this.$message({
        message: message,
        type: 'warning',
        duration: 3000
      })
    })
  },
  async mounted() {
    // 初始化时，如果有卡片则选中第一个并聚焦容器
    if (this.filteredHistory.length > 0) {
      this.selectedIndex = 0
      this.$nextTick(() => {
        this.$refs.itemsContainer?.focus()
      })
    }
    
    // 添加全局键盘事件监听
    window.addEventListener('keydown', this.handleGlobalKeydown)
  },
  beforeUnmount() {
    // 移除全局键盘事件监
    window.removeEventListener('keydown', this.handleGlobalKeydown)
  },
  watch: {
    // 监听筛选后历史记录变化
    filteredHistory: {
      handler(newHistory) {
        // 如果当前没有选中项，且有历史记录，则选中第一个并聚焦容器
        if (this.selectedIndex === -1 && newHistory.length > 0) {
          this.selectedIndex = 0
          this.$nextTick(() => {
            this.$refs.itemsContainer?.focus()
          })
        }
        // 如果选中项超出范围，重置选中状态
        else if (this.selectedIndex >= newHistory.length) {
          this.selectedIndex = newHistory.length > 0 ? 0 : -1
          if (this.selectedIndex >= 0) {
            this.$nextTick(() => {
              this.$refs.itemsContainer?.focus()
            })
          }
        }
      },
      immediate: true
    }
  },
  methods: {
    async loadHistory() {
      if (document.visibilityState === 'visible') {
        this.history = await window.go.main.App.GetHistory()
      }
    },
    async loadConfig() {
      this.config = await window.go.main.App.GetConfig()
      this.maxHistory = this.config.maxHistory
      this.tags = this.config.tags || []
      this.autoHide = this.config.autoHide || false
    },
    async copyContent(item) {
      try {
        await window.go.main.App.SaveToClipboard(item.content)
        await window.go.main.App.MoveItemToFront(item.id)
        await this.loadHistory()
        this.selectedIndex = 0
        
        this.$message({
          type: 'success',
          duration: 1000,
          showClose: false,
          customClass: 'copy-success-message'
        })
        
        if (this.autoHide) {
          window.go.main.App.HideWindow()
        }
        
      } catch (err) {
        this.$message.error('操作失败：' + err)
      }
    },
    formatDate(timestamp) {
      return new Date(timestamp).toLocaleString()
    },
    showSettings() {
      this.settingsVisible = true
    },
    async saveSettings() {
      try {
        await window.go.main.App.UpdateConfig(this.maxHistory, this.autoHide)
        this.$message.success('设置已保存')
        this.settingsVisible = false
      } catch (err) {
        this.$message.error('保存设置失败：' + err)
      }
    },
    truncateText(text, maxLength = 200) {
      if (text.length <= maxLength) return text;
      return text.substring(0, maxLength - 3) + '...';
    },
    selectItem(index) {
      this.selectedIndex = index
      // 选中时确保容器获得焦点
      this.$nextTick(() => {
        this.$refs.itemsContainer?.focus()
        this.scrollToSelectedItem()
      })
    },
    scrollToSelectedItem() {
      if (this.selectedIndex === -1) return
      
      const container = this.$refs.itemsContainer
      const items = container.children
      if (!items[this.selectedIndex]) return
      
      const selectedItem = items[this.selectedIndex]
      const containerLeft = container.scrollLeft
      const containerWidth = container.offsetWidth
      const itemLeft = selectedItem.offsetLeft
      const itemWidth = selectedItem.offsetWidth
      const margin = 20 // 边距
      
      // 立即滚动到正确位置，不使平滑滚动
      if (itemLeft + itemWidth > containerLeft + containerWidth - margin) {
        // 向右滚动时，将在容器右侧
        container.scrollLeft = itemLeft - containerWidth + itemWidth + margin
      } else if (itemLeft < containerLeft + margin) {
        // 向左滚动时，将选中项放在容器左侧
        container.scrollLeft = Math.max(0, itemLeft - margin)
      }

      // 确保容器保持焦点
      container.focus()
    },
    async deleteItem(index) {
      try {
        const item = this.filteredHistory[index]
        await window.go.main.App.DeleteHistoryItem(item.id)
        await this.loadHistory()
        
        if (this.filteredHistory.length > 0) {
          if (index >= this.filteredHistory.length) {
            this.selectedIndex = this.filteredHistory.length - 1
          } else {
            this.selectedIndex = index
          }
        } else {
          this.selectedIndex = -1
        }
        
      } catch (err) {
        this.$message.error('删除失败：' + err)
      }
    },
    handleKeyDown(event) {
      const maxIndex = this.filteredHistory.length - 1
      switch(event.key) {
        case 'ArrowLeft':
          if (this.selectedIndex > 0) {
            this.selectedIndex--
            this.scrollToSelectedItem()
          }
          break
        case 'ArrowRight':
          if (this.selectedIndex < maxIndex) {
            this.selectedIndex++
            this.scrollToSelectedItem()
          }
          break
        case 'Enter':
          if (this.selectedIndex >= 0) {
            this.copyContent(this.filteredHistory[this.selectedIndex])
          }
          break
        case 'Delete':
        case 'Backspace':
          if (this.selectedIndex >= 0) {
            this.deleteItem(this.selectedIndex)
          }
          break
      }
    },
    showAddTagDialog() {
      this.addTagVisible = true
      this.newTagName = ''
      this.newTagColor = '#409EFF'
    },
    async addTag() {
      if (!this.newTagName.trim()) {
        this.$message.warning('请入标签名称')
        return
      }
      try {
        await window.go.main.App.AddTag(this.newTagName, this.newTagColor)
        await this.loadConfig()
        this.addTagVisible = false
        this.$message.success('添加标签成功')
      } catch (err) {
        this.$message.error('添加标签失败：' + err)
      }
    },
    handleDragStart(event, item) {
      event.dataTransfer.setData('text/plain', item.id)
      event.dataTransfer.effectAllowed = 'move'
    },
    async handleDrop(event, tagId) {
      event.preventDefault()
      const itemId = event.dataTransfer.getData('text/plain')
      if (!itemId) return
      
      try {
        await window.go.main.App.UpdateItemTag(itemId, tagId)
        await this.loadHistory()
        this.$message.success('移动成功')
      } catch (err) {
        this.$message.error('移动失败：' + err)
      }
    },
    getTagColor(tagId) {
      const tag = this.tags.find(t => t.id === tagId)
      return tag ? tag.color : '#909399'
    },
    confirmDeleteTag(tag) {
      this.$confirm(`定要删除标签"${tag.name}"吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          await window.go.main.App.DeleteTag(tag.id)
          await this.loadConfig()
          if (this.currentTag === tag.id) {
            this.currentTag = null
          }
          this.$message.success('删除成功')
        } catch (err) {
          this.$message.error('删除失败：' + err)
        }
      }).catch(() => {})
    },
    toggleSearch() {
      this.showSearch = true
      this.$nextTick(() => {
        this.$refs.searchInput?.focus()
      })
    },
    handleSearchEsc() {
      this.searchText = ''
      this.showSearch = false
      // 如果有卡片，选中第一个
      if (this.filteredHistory.length > 0) {
        this.selectedIndex = 0
        this.$nextTick(() => {
          this.scrollToSelectedItem()
        })
      }
    },
    clearSearch() {
      this.searchText = ''
      // 重新聚焦搜索框
      this.$nextTick(() => {
        this.$refs.searchInput?.focus()
      })
    },
    handleGlobalKeydown(event) {
      // 如果当前有输入框在焦点中，不处理快捷键
      if (event.target.tagName === 'INPUT' || event.target.tagName === 'TEXTAREA') {
        return
      }
      
      // Command + Q 退出应用 (macOS)
      if ((event.metaKey || event.ctrlKey) && event.key === 'q') {
        event.preventDefault()
        window.go.main.App.QuitApp()
        return
      }
      
      // Command + W 隐藏窗口 (macOS)
      if ((event.metaKey || event.ctrlKey) && event.key === 'w') {
        event.preventDefault()
        window.go.main.App.HideWindow()
        return
      }
      
      // 空格键聚焦到第一个卡片
      if (event.key === ' ' || event.key === 'Spacebar') {  // 兼容不同浏览器
        event.preventDefault()
        if (this.filteredHistory.length > 0) {
          this.selectedIndex = 0
          this.$nextTick(() => {
            this.scrollToSelectedItem()
            this.$refs.itemsContainer?.focus()
          })
        }
        return
      }
      
      // Tab 键切换标签
      if (event.key === 'Tab') {
        event.preventDefault() // 阻止默认的 Tab 行为
        
        if (event.shiftKey) {
          // Shift + Tab 向左切换
          this.switchToPreviousTag()
        } else {
          // Tab 向右切换
          this.switchToNextTag()
        }
        return
      }
      
      // 按下 / 键时打开并聚焦搜索框
      if (event.key === '/') {
        event.preventDefault() // 阻止 / 字符输入到搜索框
        this.showSearch = true
        this.$nextTick(() => {
          this.$refs.searchInput?.focus()
        })
      }
    },
    showTagContextMenu(event, tag) {
      event.preventDefault()
      this.activeTag = tag
      this.showContextMenu = true
      this.contextMenuPosition = {
        x: event.clientX,
        y: event.clientY
      }
    },
    hideContextMenu() {
      this.showContextMenu = false
      this.activeTag = null
    },
    editTag() {
      this.editingTag = { ...this.activeTag }
      this.editTagVisible = true
      this.hideContextMenu()
    },
    deleteTag() {
      this.confirmDeleteTag(this.activeTag)
      this.hideContextMenu()
    },
    async saveTagEdit() {
      if (!this.editingTag.name.trim()) {
        this.$message.warning('输入标签名称')
        return
      }
      try {
        await window.go.main.App.UpdateTag(
          this.editingTag.id,
          this.editingTag.name,
          this.editingTag.color
        )
        await this.loadConfig()
        this.editTagVisible = false
        this.$message.success('编辑标签成功')
      } catch (err) {
        this.$message.error('编辑标签失败：' + err)
      }
    },
    handleTagDragStart(event, tag) {
      this.draggingTag = tag.id
      event.dataTransfer.effectAllowed = 'move'
    },
    handleTagDragEnd() {
      this.draggingTag = null
      this.dragOverTag = null
    },
    handleTagDragEnter(tag) {
      if (this.draggingTag === tag.id) return
      this.dragOverTag = tag.id
      
      // 重新排序标签
      const dragIndex = this.tags.findIndex(t => t.id === this.draggingTag)
      const dropIndex = this.tags.findIndex(t => t.id === tag.id)
      
      if (dragIndex !== -1 && dropIndex !== -1) {
        const tags = [...this.tags]
        const [draggedTag] = tags.splice(dragIndex, 1)
        tags.splice(dropIndex, 0, draggedTag)
        this.updateTagsOrder(tags)
      }
    },
    async updateTagsOrder(newTags) {
      try {
        await window.go.main.App.UpdateTagsOrder(newTags.map(tag => tag.id))
        await this.loadConfig()
      } catch (err) {
        this.$message.error('更新标签顺序失败：' + err)
      }
    },
    handleSearchEnter() {
      // 如果有搜索结果，选中第一个
      if (this.filteredHistory.length > 0) {
        this.selectedIndex = 0
        this.showSearch = false  // 隐藏搜索框
        this.$nextTick(() => {
          // 确保滚动到选中的卡片并聚焦
          this.scrollToSelectedItem()
          // 将焦点转移到卡片容器
          this.$refs.itemsContainer?.focus()
        })
      }
    },
    handleSearchInput(value) {
      // 在输入变化时保持焦点
      this.$nextTick(() => {
        this.$refs.searchInput?.focus()
      })
    },
    // 切换到下一个标签
    switchToNextTag() {
      const currentIndex = this.allTags.findIndex(tag => tag.id === this.currentTag)
      const nextIndex = (currentIndex + 1) % this.allTags.length
      this.currentTag = this.allTags[nextIndex].id
      this.selectedIndex = 0  // 重置选中的卡片到第一个
      
      // 确保标签可见
      this.$nextTick(() => {
        this.scrollTagIntoView(nextIndex)
      })
    },
    
    // 切换到上一个标签
    switchToPreviousTag() {
      const currentIndex = this.allTags.findIndex(tag => tag.id === this.currentTag)
      const prevIndex = currentIndex <= 0 ? this.allTags.length - 1 : currentIndex - 1
      this.currentTag = this.allTags[prevIndex].id
      this.selectedIndex = 0  // 重置选中的卡片到第一个
      
      // 确保标签可见
      this.$nextTick(() => {
        this.scrollTagIntoView(prevIndex)
      })
    },
    
    // 滚动标签到可见区域
    scrollTagIntoView(index) {
      const tagsContainer = this.$el.querySelector('.tags-list')
      const tagElements = tagsContainer.children
      if (tagElements[index]) {
        tagElements[index].scrollIntoView({
          behavior: 'smooth',
          block: 'nearest',
          inline: 'nearest'
        })
      }
    }
  }
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Righteous&display=swap');

:root {
  --bg-color: #f8f6f1;  /* 米白色背景 */
  --text-color: #4a4a4a;  /* 柔和的文字颜色 */
  --header-bg: #f8f6f1;
  --border-color: #e8e6e1;  /* 柔和的边框色 */
  --card-bg: #fdfcfa;  /* 略微高于背景的卡片颜色 */
  --card-hover-shadow: rgba(0, 0, 0, 0.06);
  --timestamp-border: #f0ede8;  /* 更柔和的时间戳边 */
  --dialog-bg: #ffffff;
  --dialog-text: #4a4a4a;
  --dialog-border: #e8e6e1;
  --button-hover: #f8f6f1;
}

@media (prefers-color-scheme: dark) {
  :root {
    --bg-color: #1a1a1a;
    --text-color: #ffffff;
    --header-bg: #242424;
    --border-color: #333333;
    --card-bg: #242424;
    --card-hover-shadow: rgba(255, 255, 255, 0.05);
    --timestamp-border: #333333;
    --dialog-bg: #242424;
    --dialog-text: #ffffff;
    --dialog-border: #333333;
    --button-hover: #1a1a1a;
  }
}

.app-root {
  width: 100vw;
  height: 100vh;
  background-color: var(--bg-color);
  overflow: hidden;
}

html, body {
  margin: 0;
  padding: 0;
  background-color: var(--bg-color);
  overflow: hidden;
}

.container {
  background-color: var(--bg-color);
  color: var(--text-color);
  height: 100%;
  overflow: hidden;
}

.el-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: transparent;
  padding: 0 20px;
  height: 50px;
  color: var(--text-color);
}

.el-header h2 {
  margin: 0;
  font-weight: 500;
}

.el-header .el-button {
  padding: 0;
  margin-left: 10px;
}

.clipboard-items {
  display: flex;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 10px 4px;
  scrollbar-width: none;
  -ms-overflow-style: none;
  height: 100%;
  scroll-behavior: auto;
  padding-top: 12px;
  padding-bottom: 20px;
}

.clipboard-items::-webkit-scrollbar {
  display: none;
}

.clipboard-item {
  flex: 0 0 auto;
  width: 180px;
  height: 180px;
  margin-right: 20px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  background-color: var(--card-bg);
  position: relative;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transform: translateY(0);
}

.clipboard-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(149, 157, 165, 0.1);
}

.clipboard-item.selected {
  transform: translateY(-6px);
  border-color: #9b59b6;
  box-shadow: 0 0 0 2px rgba(155, 89, 182, 0.2), 0 8px 24px rgba(155, 89, 182, 0.15);
}

.clipboard-image {
  max-width: 100%;
  max-height: 140px;
  object-fit: contain;
  margin-bottom: 5px;
}

.text-content {
  word-break: break-word;
  overflow: hidden;
  max-height: 140px;
  display: -webkit-box;
  -webkit-line-clamp: 7;
  -webkit-box-orient: vertical;
  font-size: 14px;
  line-height: 1.4;
  margin-bottom: 5px;
  color: var(--text-color);
}

.item-timestamp {
  font-size: 12px;
  color: var(--text-color);
  opacity: 0.6;
  margin-top: auto;
  padding-top: 5px;
  text-align: center;
  width: 100%;
  border-top: 1px solid var(--timestamp-border);
}

.el-main {
  padding: 10px;
  padding-right: 14px;
  height: calc(100vh - 90px);
  overflow: hidden;
  background-color: var(--bg-color);
}

.clipboard-items:focus {
  outline: none;
}

/* 添加不支持内容的样式 */
.unsupported-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 130px;
  color: #e6a23c;
  gap: 8px;
}

.unsupported-content .el-icon {
  font-size: 24px;
}

/* 修改数字输入框样式 */
.max-history-input {
  width: 120px !important;
}

.max-history-input .el-input-number__decrease,
.max-history-input .el-input-number__increase {
  width: 32px !important;
  height: 100% !important;
  top: 0px !important;
}

.max-history-input .el-input-number__decrease {
  left: 0px !important;
  border-radius: 4px 0 0 4px !important;
}

.max-history-input .el-input-number__increase {
  right: 0px !important;
  border-radius: 0 4px 4px 0 !important;
}

.max-history-input .el-input__wrapper {
  padding: 0 32px !important;
  box-shadow: 0 0 0 1px var(--border-color) !important;
}

.max-history-input .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px var(--border-color) !important;
}

.max-history-input .el-input__inner {
  text-align: center !important;
  padding: 0 !important;
  height: 32px !important;
  line-height: 32px !important;
}

/* 调整对话框样式 */
.el-dialog {
  margin-top: 15vh !important;
  background-color: var(--dialog-bg) !important;
  border: 1px solid var(--dialog-border) !important;
}

.el-dialog__title {
  color: var(--dialog-text) !important;
}

.el-dialog__body {
  color: var(--dialog-text) !important;
}

.el-form-item {
  margin-bottom: 0 !important;
}

.el-form-item__label {
  color: var(--dialog-text) !important;
}

.app-title {
  margin: 0;
  font-family: 'Righteous', cursive;
  font-size: 24px;
  background: linear-gradient(135deg, #9b59b6 0%, #3498db 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-fill-color: transparent;
  letter-spacing: 1px;
  text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
  user-select: none;
  -webkit-user-select: none;
  -moz-user-select: none;
  -ms-user-select: none;
}

/* 暗黑模式下的标题效果 */
@media (prefers-color-scheme: dark) {
  .app-title {
    background: linear-gradient(135deg, #9b59b6 0%, #3498db 100%);
    -webkit-background-clip: text;
    background-clip: text;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.2);
  }
}

/* 对话框按样式 */
.dialog-footer .el-button {
  background-color: var(--dialog-bg) !important;
  border-color: var(--dialog-border) !important;
  color: var(--dialog-text) !important;
}

.dialog-footer .el-button:hover {
  background-color: var(--button-hover) !important;
}

.dialog-footer .el-button--primary {
  background-color: #9b59b6 !important;
  border-color: #9b59b6 !important;
  color: #ffffff !important;
}

.dialog-footer .el-button--primary:hover {
  background-color: #8e44ad !important;
  border-color: #8e44ad !important;
}

/* 数字输入框暗黑模式样式 */
@media (prefers-color-scheme: dark) {
  .max-history-input .el-input__wrapper {
    background-color: var(--dialog-bg) !important;
  }
  
  .max-history-input .el-input__inner {
    color: var(--dialog-text) !important;
  }
  
  .max-history-input .el-input-number__decrease,
  .max-history-input .el-input-number__increase {
    background-color: var(--dialog-bg) !important;
    color: var(--dialog-text) !important;
    border-color: var(--dialog-border) !important;
  }
  
  .max-history-input .el-input-number__decrease:hover,
  .max-history-input .el-input-number__increase:hover {
    background-color: var(--button-hover) !important;
  }
}

.tags-bar {
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 40px;
  border-bottom: 1px solid var(--border-color);
  background-color: transparent;
  gap: 8px;
}

.tags-list {
  display: flex;
  flex: 1;
  overflow-x: auto;
  gap: 8px;
  scrollbar-width: none;
  -ms-overflow-style: none;
  padding: 4px 0;
  background-color: transparent;
  padding-top: 1px;
  padding-bottom: 1px;
}

.tag-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  border-radius: 12px;
  background-color: transparent;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s ease;
  border: 1px solid var(--border-color);
  font-size: 12px;
  letter-spacing: 0.3px;
  color: var(--text-color);
  opacity: 0.8;
}

.tag-item:hover {
  opacity: 1;
  border-color: #9b59b6;
}

.tag-item.active {
  background-color: rgba(155, 89, 182, 0.05);
  border-color: #9b59b6;
  opacity: 1;
  color: var(--text-color);
}

.tag-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 8px;
  transition: transform 0.2s ease;
}

.tag-item:hover .tag-dot {
  transform: scale(1.2);
}

.add-tag-btn {
  margin-left: 8px;
  color: var(--text-color);
  opacity: 0.6;
  transition: all 0.2s ease;
}

.add-tag-btn:hover {
  opacity: 1;
  color: #9b59b6;
  transform: rotate(90deg);
}

/* 暗黑模式下的标签样式调整 */
@media (prefers-color-scheme: dark) {
  .tag-item {
    background-color: transparent;
    border-color: var(--border-color);
  }

  .tag-item:hover {
    border-color: #9b59b6;
    background-color: rgba(155, 89, 182, 0.1);
  }

  .tag-item.active {
    background-color: rgba(155, 89, 182, 0.15);
    border-color: #9b59b6;
  }
}

.item-tag-indicator {
  position: absolute;
  top: 0;
  left: 0;
  width: 4px;
  height: 100%;
  border-radius: 4px 0 0 4px;
}

.delete-tag-icon {
  display: none;
}

.add-tag-btn {
  margin-left: 8px;
  color: var(--text-color);
  opacity: 0.6;
  transition: all 0.2s ease;
}

.add-tag-btn:hover {
  opacity: 1;
  color: #9b59b6;
  transform: rotate(90deg);
}

/* 添加标签输入组样式 */
.tag-input-group {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: var(--dialog-bg);
  border: 1px solid var(--dialog-border);
  border-radius: 4px;
  padding: 2px;
}

.tag-color-picker {
  margin: 0 4px;
}

.tag-name-input {
  flex: 1;
}

/* 输入框暗黑模式适配 */
.tag-name-input .el-input__wrapper {
  box-shadow: none !important;
  padding-left: 0 !important;
  background-color: var(--dialog-bg) !important;
}

.tag-name-input .el-input__wrapper:hover {
  box-shadow: none !important;
}

.tag-name-input .el-input__inner {
  color: var(--dialog-text) !important;
  background-color: var(--dialog-bg) !important;
}

/* 暗黑模式下的输入框样式 */
@media (prefers-color-scheme: dark) {
  .tag-input-group {
    background-color: var(--dialog-bg);
    border-color: var(--dialog-border);
  }
  
  .tag-name-input .el-input__wrapper {
    background-color: var(--dialog-bg) !important;
  }
  
  .tag-name-input .el-input__inner {
    color: var(--dialog-text) !important;
    background-color: var(--dialog-bg) !important;
  }
  
  /* 颜色选择器暗黑模式适配 */
  .el-color-picker__trigger {
    border-color: var(--dialog-border) !important;
    background-color: var(--dialog-bg) !important;
  }
}

.search-container {
  display: flex;
  align-items: center;
}

.search-btn {
  color: var(--text-color);
  opacity: 0.6;
  transition: all 0.2s ease;
}

.search-btn:hover {
  opacity: 1;
  color: #9b59b6;
}

.search-input-container {
  width: 200px;
  transition: all 0.3s ease;
}

.search-input {
  font-size: 13px;
}

.search-input .el-input__wrapper {
  background-color: transparent !important;
  box-shadow: 0 0 0 1px var(--border-color) !important;
  border-radius: 12px !important;
  padding: 0 12px !important;
  height: 28px !important;
}

.search-input .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px #9b59b6 !important;
}

.search-input .el-input__inner {
  color: var(--text-color) !important;
  height: 28px !important;
  line-height: 28px !important;
  font-size: 13px !important;
}

.search-input .el-input__prefix-inner {
  color: var(--text-color);
  opacity: 0.6;
  margin-right: 4px;
}

.search-input .el-input__suffix-inner {
  margin-left: 4px;
}

.clear-search {
  cursor: pointer;
  opacity: 0.6;
  transition: all 0.2s ease;
}

.clear-search:hover {
  opacity: 1;
  color: #9b59b6;
}

/* 暗黑模式适配 */
@media (prefers-color-scheme: dark) {
  .search-input .el-input__wrapper {
    background-color: var(--dialog-bg) !important;
  }
  
  .search-input .el-input__inner::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }
}

/* 添加悬浮菜单样式 */
.tag-popup-menu {
  position: fixed;
  background: var(--dialog-bg);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 4px 0;
  min-width: 100px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  z-index: 9999;
}

.menu-item {
  padding: 6px 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-color);
  font-size: 13px;
}

.menu-item:hover {
  background: var(--button-hover);
}

.menu-item.delete {
  color: #f56c6c;
}

.menu-item .el-icon {
  font-size: 14px;
}

/* 黑模式配 */
@media (prefers-color-scheme: dark) {
  .tag-popup-menu {
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.3);
  }
}

/* 移除原来的右键菜单样式 */
.tag-context-menu {
  display: none;
}

.empty-state {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-color);
  opacity: 0.4;
  user-select: none;
  transform: scale(0.95);
  transition: all 0.3s ease;
}

.empty-state:hover {
  opacity: 0.6;
  transform: scale(1);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 14px;
  margin: 0;
}

.tag-item.all-tag {
  cursor: pointer;
  background-color: transparent;
  border-color: var(--border-color);
}

.tag-item.all-tag:hover {
  opacity: 1;
  border-color: #9b59b6;
}

.tag-item.all-tag.active {
  background-color: rgba(155, 89, 182, 0.05);
  border-color: #9b59b6;
  opacity: 1;
  color: var(--text-color);
}

/* 暗黑模式下的全部标签样式 */
@media (prefers-color-scheme: dark) {
  .tag-item.all-tag.active {
    background-color: rgba(155, 89, 182, 0.15);
    border-color: #9b59b6;
  }
}

.tag-item.dragging {
  opacity: 0.5;
  transform: scale(0.95);
  border-color: transparent !important;
  box-shadow: none !important;
  background-color: transparent !important;
}

.tag-item.dragging:hover {
  border-color: transparent !important;
  box-shadow: none !important;
  transform: scale(0.95) !important;
  background-color: transparent !important;
}

.tag-item.dragging.active {
  border-color: transparent !important;
  box-shadow: none !important;
  background-color: transparent !important;
}

/* 暗黑模式下的拖拽样式 */
@media (prefers-color-scheme: dark) {
  .tag-item.dragging,
  .tag-item.dragging:hover,
  .tag-item.dragging.active {
    border-color: transparent !important;
    background-color: transparent !important;
    box-shadow: none !important;
  }
}

/* 自定义消息提示样式 */
.el-message {
  min-width: auto !important;
  padding: 8px 12px !important;
}

.el-message--success {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.el-message__content {
  display: none !important;  /* 隐藏文字内容 */
}

.el-message .el-message__icon {
  margin-right: 0 !important;  /* 移除图标右边距 */
  font-size: 18px !important;  /* 调整图标大小 */
}

/* 自定义消息提示样式 */
.copy-success-message {
  min-width: auto !important;
  padding: 8px !important;
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
}

.copy-success-message .el-message__content {
  display: none !important;
}

.copy-success-message .el-message__icon {
  margin: 0 !important;
  font-size: 20px !important;
  color: #67c23a !important;
}

/* 添加开关样式 */
.el-form-item {
  margin-bottom: 20px !important;
}

.clipboard-items::-webkit-scrollbar {
  height: 6px;
}

.clipboard-items::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 3px;
}

.clipboard-items::-webkit-scrollbar-track {
  background: transparent;
}
</style> 