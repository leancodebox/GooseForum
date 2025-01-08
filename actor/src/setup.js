import { createApp } from 'vue'
import {
    create,
    NButton,
    NCard,
    NForm,
    NFormItem,
    NInput,
    NResult,
    NStep,
    NSteps,
    NMessageProvider,
} from 'naive-ui'
import Setup from './Setup.vue'

const naive = create({
    components: [
        NButton,
        NCard,
        NForm,
        NFormItem,
        NInput,
        NResult,
        NStep,
        NSteps,
    ]
})

const app = createApp(Setup)
app.use(naive)
app.mount('#app')
