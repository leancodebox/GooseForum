import './gf-style.css'
import Alpine from 'alpinejs'
import axios from 'axios'
import * as api from './utils/gooseForumService'
import {renderMermaidDiagrams} from "@/utils/mermaidRenderer.ts";
import { formatTimeAgoStr, formatNumberStr } from './utils/formatters'

// Expose libraries
(window as any).Alpine = Alpine;
(window as any).axios = axios;
(window as any).api = api;
(window as any).MermaidUtils = { renderMermaidDiagrams };
(window as any).formatTimeAgoStr = formatTimeAgoStr;
(window as any).formatNumberStr = formatNumberStr;

// Lazy load ImageUtils
(window as any).ImageUtils = {
    processImageFile: async (file: File, quality?: number, onProgress?: (message: string) => void) => {
        const module = await import('./utils/imageUtils');
        return module.processImageFile(file, quality, onProgress);
    }
};

// Lazy load Cropper
(window as any).loadCropper = async () => {
    // @ts-ignore
    await import('cropperjs/dist/cropper.css');
    const module = await import('cropperjs');
    (window as any).Cropper = module.default;
    return module.default;
};

// V2 Specific Logic
document.addEventListener('alpine:init', () => {
    Alpine.data('homeV2', () => ({
        init() {
            console.log('Home V2 Initialized');
        }
    }))

    Alpine.data('followUserHandler', (userId, initialStatus = false) => ({
        isFollowing: initialStatus,
        loading: false,
        init() {
        },
        async checkStatus() {
        },
        async toggleFollow() {
            if (this.loading) return;
            this.loading = true;
            try {
                const action = this.isFollowing ? 2 : 1;
                const response = await api.followUser(userId, action);
                if (response.code === 0) {
                     this.isFollowing = !this.isFollowing;
                }
            } catch (error) {
                console.error('Follow action error:', error);
            } finally {
                this.loading = false;
            }
        }
    }))
})
Alpine.start()
