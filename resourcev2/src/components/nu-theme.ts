
import { NConfigProvider, GlobalThemeOverrides } from 'naive-ui'


const themeOverrides: GlobalThemeOverrides = {
    Select: {
      peers: {
        InternalSelection: {
          textColor: '#FF0000'
        },
        InternalSelectMenu: {
          borderRadius: '6px'
        }
      }
    }
}