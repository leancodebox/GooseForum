import './style-v2.css'
import Alpine from 'alpinejs'
import axios from 'axios'

// Expose libraries
(window as any).Alpine = Alpine;
(window as any).axios = axios;

// V2 Specific Logic
document.addEventListener('alpine:init', () => {
    Alpine.data('homeV2', () => ({
        init() {
            console.log('Home V2 Initialized');
        }
    }))
})

Alpine.start()
