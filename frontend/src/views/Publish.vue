<template>
  <div class="publish-page">
    <nav class="navbar">
      <div class="container navbar-content">
        <div class="brand" @click="$router.push('/')">
          <div class="logo-box">
            <div class="circle-shape"></div>
            <div class="square-shape"></div>
          </div>
          <span class="brand-text">XIANQU</span>
        </div>
        <div class="nav-title">
          {{ isEditMode ? '编辑闲置' : '发布闲置' }}
        </div>
        <div class="nav-actions">
           <button class="close-btn" @click="$router.go(-1)">取消</button>
        </div>
      </div>
    </nav>

    <main class="main-container" v-loading="loading">
      <div class="publish-content">
        <div class="media-section card animate-fade">
          <h3 class="section-title">商品图片</h3>
          <p class="section-subtitle">上传一张好看的封面图，吸引更多眼球</p>
          <el-upload
            class="image-uploader"
            :action="uploadUrl"
            name="file"
            :headers="uploadHeaders"
            :show-file-list="false"
            :on-success="handleUploadSuccess"
            :before-upload="beforeUpload"
            :on-error="handleUploadError"
          >
            <div class="upload-area" :class="{ 'has-image': publishForm.image }">
              <img v-if="publishForm.image" :src="publishForm.image" class="preview-img" />
              <div v-else class="upload-placeholder">
                <el-icon class="upload-icon"><Plus /></el-icon>
                <div class="upload-text">点击上传封面图</div>
              </div>
              <div class="hover-mask" v-if="publishForm.image">
                <el-icon><Switch /></el-icon> 更换图片
              </div>
            </div>
          </el-upload>
        </div>

        <div class="form-section card animate-fade delay-1">
          <div class="form-group margin-b">
            <input
              v-model="publishForm.name"
              class="custom-input title-input"
              placeholder="商品名称、品牌型号..."
              maxlength="50"
            />
          </div>
          <div class="form-group margin-b">
            <textarea
              v-model="publishForm.description"
              class="custom-textarea"
              rows="4"
              placeholder="描述一下宝贝的入手渠道、新旧程度和转手原因..."
            ></textarea>
          </div>

          <div class="form-row margin-b">
            <div class="form-group half">
              <label>价格 (¥)</label>
              <input
                v-model.number="publishForm.price"
                type="number"
                class="custom-input price-input"
                placeholder="0.00"
              />
            </div>
            <div class="form-group half">
              <label>数量</label>
              <el-input-number
                v-model="publishForm.count"
                :min="1"
                :max="99"
                controls-position="right"
                class="custom-number"
              />
            </div>
          </div>

          <div class="form-group margin-b">
            <label>商品分类</label>
            <div class="category-grid">
               <div v-for="cat in categories" :key="cat.id" 
                    class="category-tag" 
                    :class="{ active: publishForm.category === cat.name }"
                    @click="publishForm.category = cat.name">
                 {{ cat.name }}
               </div>
            </div>
          </div>

          <div class="form-group margin-b">
            <label>发货地</label>
            <div class="location-wrapper">
              <input 
                v-model="publishForm.area" 
                class="custom-input" 
                placeholder="所在区域 (例如：清华大学)" 
              />
              <button class="locate-btn" @click="getLocation" type="button">
                <el-icon><Aim /></el-icon> 定位
              </button>
            </div>
            <div v-if="locationErrorMsg" class="error-text">{{ locationErrorMsg }}</div>
          </div>

          <div class="form-group margin-b">
            <label>交易选项</label>
            <div class="trade-options">
               <div class="trade-option" :class="{ active: publishForm.is_negotiable }" @click="toggleTradeOption('is_negotiable')">
                 <span class="opt-left">
                   <el-icon v-if="publishForm.is_negotiable"><Select /></el-icon>
                   <el-icon v-else><Scissor /></el-icon>可小刀
                 </span>
                 <el-switch v-model="publishForm.is_negotiable" @click.stop />
               </div>
               <div class="trade-option" :class="{ active: publishForm.is_home_delivery }" @click="toggleTradeOption('is_home_delivery')">
                 <span class="opt-left">
                   <el-icon v-if="publishForm.is_home_delivery"><Select /></el-icon>
                   <el-icon v-else><Van /></el-icon>送货上门
                 </span>
                 <el-switch v-model="publishForm.is_home_delivery" @click.stop />
               </div>
               <div class="trade-option" :class="{ active: publishForm.is_self_pickup }" @click="toggleTradeOption('is_self_pickup')">
                 <span class="opt-left">
                   <el-icon v-if="publishForm.is_self_pickup"><Select /></el-icon>
                   <el-icon v-else><Goods /></el-icon>自提
                 </span>
                 <el-switch v-model="publishForm.is_self_pickup" @click.stop />
               </div>
            </div>
          </div>

          <div class="action-footer">
            <div class="status-box" v-if="isEditMode">
              <el-radio-group v-model="publishForm.status">
                <el-radio :label="1">上架中</el-radio>
                <el-radio :label="3">已下架</el-radio>
                <el-radio :label="2" disabled v-if="publishForm.status === 2">已售出</el-radio>
              </el-radio-group>
            </div>
            <div class="spacer" v-else></div>

            <button class="submit-btn" @click="submitPublish" :disabled="isSubmitting">
              {{ isSubmitting ? '提交中...' : (isEditMode ? '保存修改' : '确认发布') }}
            </button>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import request from '@/utils/request'
import { ElMessage, ElLoading } from 'element-plus'
import { Plus, Switch, Van, Goods, Scissor, Close, Location, Aim, Select } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const isSubmitting = ref(false)
const loading = ref(false)
const locationErrorMsg = ref('')
const user = JSON.parse(localStorage.getItem('user') || '{}')

const isEditMode = computed(() => !!route.query.id)

const uploadUrl = computed(() => {
  const baseUrl = import.meta.env.VITE_API_URL || '/'
  return baseUrl.endsWith('/') ? `${baseUrl}api/upload` : `${baseUrl}/api/upload`
})

const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token')
  return { Authorization: token || '' }
})

const categories = ref([
  { id: 1, name: '数码', group: '电子设备' },
  { id: 2, name: '手游', group: '电子设备' },
  { id: 3, name: '书籍', group: '学习生活' },
  { id: 4, name: '生活', group: '学习生活' },
  { id: 5, name: '兼职', group: '学习生活' },
  { id: 6, name: '服饰', group: '穿搭美学' },
  { id: 7, name: '运动', group: '穿搭美学' },
  { id: 8, name: '美妆', group: '穿搭美学' },
  { id: 9, name: '乐器', group: '兴趣收藏' },
  { id: 10, name: '其他', group: '兴趣收藏' }
])

const categoryGroups = computed(() => {
  const groupMap = new Map()
  categories.value.forEach((cat) => {
    if (!groupMap.has(cat.group)) groupMap.set(cat.group, [])
    groupMap.get(cat.group).push(cat)
  })

  return Array.from(groupMap.entries()).map(([key, items]) => ({ key, items }))
})

const selectedTradeOptions = computed(() => {
  const options = []
  if (publishForm.is_negotiable) options.push('可小刀')
  if (publishForm.is_home_delivery) options.push('送货上门')
  if (publishForm.is_self_pickup) options.push('自提')
  return options
})

const toBool = (value) => {
  return value === true || value === 1 || value === '1' || value === 'true'
}

const publishForm = reactive({
  name: '',
  description: '',
  price: '',
  count: 1,
  image: '',
  category: '',
  area: '',
  user_id: user.id,
  status: 1,
  is_negotiable: false,
  is_home_delivery: false,
  is_self_pickup: false
})

const getLocation = () => {
  locationErrorMsg.value = ''

  if (typeof AMap === 'undefined') {
    locationErrorMsg.value = '地图资源未加载，请刷新页面重试'
    return ElMessage.error('地图SDK未加载')
  }

  const loadingInstance = ElLoading.service({ text: '正在定位中...', background: 'rgba(0,0,0,0.7)' })

  AMap.plugin(['AMap.Geolocation', 'AMap.Geocoder'], function() {
    var geolocation = new AMap.Geolocation({
      enableHighAccuracy: true,
      timeout: 10000,
      extensions: 'base'
    });

    geolocation.getCurrentPosition(function(status, result) {
      if (status === 'complete') {
        const position = result.position;
        const geocoder = new AMap.Geocoder({ radius: 500, extensions: 'all' });

        geocoder.getAddress([position.lng, position.lat], function(status, regeoResult) {
          loadingInstance.close()

          if (status === 'complete' && regeoResult.regeocode) {
            const regeo = regeoResult.regeocode;
            const pois = regeo.pois || [];
            let bestName = '';

            for (let poi of pois) {
              const type = poi.type || '';
              const name = poi.name || '';
              if (type.includes('学校') || type.includes('高等院校') || type.includes('中学') || type.includes('大学')) {
                bestName = name; break;
              }
              if (type.includes('住宅') || type.includes('居住') || type.includes('小区') || type.includes('公寓')) {
                if (!bestName) bestName = name;
              }
            }
            if (!bestName) {
              bestName = regeo.addressComponent.neighborhood || regeo.addressComponent.building || regeo.addressComponent.township || regeo.formattedAddress;
            }

            publishForm.area = bestName;
            ElMessage.success(`已定位到：${bestName}`);
            localStorage.setItem('userLocation', bestName);
          } else {
            publishForm.area = result.formattedAddress;
            ElMessage.warning('未能精确匹配小区，已显示大致位置');
          }
        });
      } else {
        loadingInstance.close()
        console.error('定位失败:', result)
        let msg = '定位失败'
        if (result.message && result.message.includes('permission denied')) {
          msg = '请允许浏览器获取位置权限'
        } else if (result.info === 'FAILED') {
          msg = '定位超时或网络不稳定，请手动输入'
        }
        locationErrorMsg.value = `${msg} (${result.message || result.info})`
        ElMessage.error(msg)
      }
    });
  });
}

const fetchProductData = async () => {
  if (!isEditMode.value) return

  loading.value = true
  try {
    const res = await request.get(`/api/products/${route.query.id}`)
    const data = res.data

    if (data.user_id !== user.id) {
      ElMessage.error('无权编辑此商品')
      router.push('/')
      return
    }

    Object.assign(publishForm, {
      name: data.name,
      description: data.description,
      price: data.price,
      count: data.count || 1,
      category: data.category,
      image: data.image,
      area: data.area || '',
      status: data.status,
      is_negotiable: toBool(data.is_negotiable ?? data.isNegotiable),
      is_home_delivery: toBool(data.is_home_delivery ?? data.isHomeDelivery),
      is_self_pickup: toBool(data.is_self_pickup ?? data.isSelfPickup)
    })

  } catch (e) {
    console.error(e)
    ElMessage.error('获取商品信息失败')
  } finally {
    loading.value = false
  }
}

const handleUploadSuccess = (res) => {
  if (res.url) {
    let finalUrl = res.url
    publishForm.image = finalUrl
    ElMessage.success('图片上传成功')
  }
}

const handleUploadError = (err) => {
  console.error('Upload failed:', err)
  ElMessage.error('上传失败，请检查网络')
}

const beforeUpload = (file) => {
  const isLt100M = file.size / 1024 / 1024 < 100
  if (!isLt100M) ElMessage.warning('图片大小不能超过 100MB')
  return isLt100M
}

const toggleTradeOption = (field) => {
  publishForm[field] = !publishForm[field]
}

const submitPublish = async () => {
  if (!publishForm.name) return ElMessage.warning('请输入标题')
  if (!publishForm.price) return ElMessage.warning('请输入价格')
  if (!publishForm.image) return ElMessage.warning('请上传图片')
  if (!publishForm.category) return ElMessage.warning('请选择分类')

  isSubmitting.value = true
  try {
    const submitData = {
      name: publishForm.name,
      description: publishForm.description,
      price: Number(publishForm.price),
      count: Number(publishForm.count),
      image: publishForm.image,
      category: publishForm.category,
      area: publishForm.area,
      status: publishForm.status,
      is_negotiable: publishForm.is_negotiable,
      is_home_delivery: publishForm.is_home_delivery,
      is_self_pickup: publishForm.is_self_pickup,
      // 兼容历史后端字段绑定（camelCase）
      isNegotiable: publishForm.is_negotiable,
      isHomeDelivery: publishForm.is_home_delivery,
      isSelfPickup: publishForm.is_self_pickup
    }

    if (isEditMode.value) {
      await request.put(`/api/products/${route.query.id}`, submitData)
      ElMessage.success('修改成功！')
    } else {
      await request.post('/api/products', submitData)
      ElMessage.success('发布成功！')
    }

    setTimeout(() => {
      if (isEditMode.value) {
        router.push('/profile')
      } else {
        router.push('/')
      }
    }, 1000)

  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.error || (isEditMode.value ? '修改失败' : '发布失败'))
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchProductData()
  if (!isEditMode.value && !publishForm.area) {
    const cachedLoc = localStorage.getItem('userLocation')
    if (cachedLoc) {
      publishForm.area = cachedLoc
    }
  }
})
</script>

<style scoped lang="scss">
$primary: #ffda44;
$bg-color: #f6f7f9;

.publish-page {
  min-height: 100vh;
  background: $bg-color;
  font-family: "PingFang SC", -apple-system, sans-serif;
  color: #333;
}

.navbar {
  height: 64px;
  background: $primary;
  position: sticky;
  top: 0;
  z-index: 100;
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);

  .navbar-content {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
  }
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  
  .logo-box {
    width: 32px;
    height: 32px;
    position: relative;
    .circle-shape { position: absolute; width: 20px; height: 20px; background: #000; border-radius: 50%; top: 0; left: 0; }
    .square-shape { position: absolute; width: 18px; height: 18px; border: 2px solid #fff; bottom: 0; right: 0; border-radius: 4px; background: rgba(255,255,255,0.2); backdrop-filter: blur(2px); }
  }
  
  .brand-text {
    font-size: 22px;
    font-weight: 900;
    color: #000;
    letter-spacing: 2px;
    font-family: 'Arial Black', sans-serif;
  }
}

.nav-title {
  font-size: 18px;
  font-weight: 800;
  color: #000;
}

.close-btn {
  background: #000;
  color: $primary;
  border: none;
  border-radius: 99px;
  padding: 6px 20px;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 14px;
  
  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  }
}

.main-container {
  max-width: 1200px;
  margin: 40px auto;
  padding: 0 20px 80px;
}

.publish-content {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.card {
  background: #fff;
  border-radius: 20px;
  padding: 32px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.02);
}

// Optimized Media Section
.media-section {
  flex: 0 0 380px;
  position: sticky;
  top: 104px;
  display: flex;
  flex-direction: column;
}

.form-section {
  flex: 1;
  min-width: 0;
}

.section-title {
  margin: 0 0 8px;
  font-size: 20px;
  font-weight: 800;
  color: #000;
}
.section-subtitle {
  margin: 0 0 28px;
  font-size: 14px;
  color: #999;
  line-height: 1.4;
}

.upload-area {
  width: 100%;
  aspect-ratio: 1;
  border-radius: 16px;
  border: 2px dashed #e5e7eb;
  background: #fdfdfd;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  
  &:hover {
    border-color: $primary;
    background: #fff;
    transform: scale(1.01);
  }
  
  &.has-image {
    border: none;
    background: #000;
    box-shadow: 0 10px 25px rgba(0,0,0,0.1);
  }
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover; /* Changed to cover for a more full look */
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #999;
  
  .upload-icon {
    font-size: 42px;
    color: #000;
    background: $primary;
    padding: 16px;
    border-radius: 50%;
    box-shadow: 0 4px 12px rgba(255, 218, 68, 0.3);
  }
  
  .upload-text {
    font-weight: 600;
    font-size: 15px;
    color: #666;
  }
}

.hover-mask {
  position: absolute;
  inset: 0;
  background: rgba(0,0,0,0.6);
  backdrop-filter: blur(4px);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  font-weight: 700;
  opacity: 0;
  transition: all 0.3s;
}

.upload-area:hover .hover-mask {
  opacity: 1;
}

.margin-b { margin-bottom: 28px; }

.form-row { 
  display: flex; 
  gap: 24px; 
}
.half { flex: 1; }

.form-group {
  display: flex;
  flex-direction: column;
  
  label {
    font-size: 14px;
    font-weight: 700;
    color: #333;
    margin-bottom: 12px;
    padding-left: 4px;
  }
}

.custom-input, .custom-textarea {
  width: 100%;
  box-sizing: border-box;
  border-radius: 14px;
  padding: 15px 20px;
  border: 1px solid transparent;
  background: #f5f6f8;
  font-size: 15px;
  color: #000;
  transition: all 0.25s ease;
  font-family: inherit;
  
  &::placeholder {
    color: #aaa;
  }
  
  &:focus {
    outline: none;
    background: #fff;
    border-color: $primary;
    box-shadow: 0 0 0 4px rgba(255, 218, 68, 0.1);
  }
}

.title-input {
  font-size: 22px;
  font-weight: 800;
  background: #fff;
  border-bottom: 2px solid #f0f0f0;
  border-radius: 0;
  padding: 12px 0;
  margin-bottom: 10px;
  
  &:focus {
    border-bottom-color: $primary;
    box-shadow: none;
  }
}

.price-input {
  font-size: 20px;
  font-weight: 800;
  color: #ff5000;
  background: #fff6f6;
  border: 1px solid #ffecec;
  
  &:focus {
    background: #fff;
    border-color: #ff5000;
    box-shadow: 0 0 0 4px rgba(255, 80, 0, 0.05);
  }
}

.custom-textarea {
  resize: none;
  min-height: 160px;
  line-height: 1.6;
}

.custom-number {
  width: 100%;
  :deep(.el-input__wrapper) {
    box-shadow: none !important;
    border: 1px solid transparent;
    border-radius: 14px;
    background: #f5f6f8;
    height: 52px;
    transition: all 0.25s;
    &.is-focus {
      background: #fff;
      border-color: $primary !important;
      box-shadow: 0 0 0 4px rgba(255, 218, 68, 0.1) !important;
    }
  }
}

.category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.category-tag {
  padding: 10px 22px;
  border-radius: 12px;
  background: #f0f2f5;
  color: #666;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  
  &:hover {
    background: #e5e7eb;
    color: #000;
    transform: translateY(-1px);
  }
  
  &.active {
    background: #000;
    color: $primary;
    transform: translateY(-2px);
    box-shadow: 0 8px 16px rgba(0,0,0,0.12);
  }
}

.location-wrapper {
  display: flex;
  gap: 12px;
  
  .locate-btn {
    border: none;
    background: #000;
    color: $primary;
    border-radius: 12px;
    padding: 0 24px;
    font-weight: 700;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;
    transition: all 0.3s;
    
    &:hover {
      background: #222;
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    }
  }
}

.trade-options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 16px;
}

.trade-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-radius: 16px;
  background: #fff;
  border: 1px solid #f0f0f0;
  cursor: pointer;
  transition: all 0.3s;
  
  .opt-left {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 14px;
    font-weight: 600;
    color: #555;
  }
  
  &:hover {
    border-color: #ddd;
    background: #fafafa;
  }
  
  &.active {
    background: #fffdf5;
    border-color: $primary;
    box-shadow: 0 4px 12px rgba(255, 218, 68, 0.1);
    .opt-left { color: #000; font-weight: 700; }
  }
}

.action-footer {
  margin-top: 48px;
  padding-top: 32px;
  border-top: 2px solid #f6f6f6;
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 24px;
}

.submit-btn {
  background: #000;
  color: $primary;
  border: none;
  border-radius: 16px;
  padding: 18px 80px;
  font-size: 18px;
  font-weight: 900;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 12px 24px rgba(0,0,0,0.12);
  
  &:hover {
    transform: translateY(-4px);
    box-shadow: 0 16px 32px rgba(0,0,0,0.18);
  }
  
  &:active {
    transform: translateY(-1px);
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none;
  }
}

@media (max-width: 1024px) {
  .media-section { flex: 0 0 320px; }
}

@media (max-width: 850px) {
  .publish-content { flex-direction: column; }
  .media-section { position: static; width: 100%; flex: none; }
  .nav-title { display: none; }
}

.animate-fade {
  animation: fadeIn 0.6s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  opacity: 0;
  transform: translateY(20px);
}
.delay-1 { animation-delay: 0.15s; }

@keyframes fadeIn {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
