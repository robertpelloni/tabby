import * as React from 'react'
import * as ReactDOM from 'react-dom/client'

// Use the global registry injected by ReactPluginDecorator in tabby-terminal
const registry = (window as any)['tabbyReactPlugins']

if (registry) {
    registry.register((tab: any) => {
        // Create an element that stays mounted above the terminal tab
        const wrapper = document.createElement('div')
        wrapper.className = 'tabby-react-dummy-plugin'
        wrapper.style.position = 'absolute'
        wrapper.style.top = '10px'
        wrapper.style.right = '50px' // Leave room for the scrollbar/actions
        wrapper.style.padding = '5px 10px'
        wrapper.style.background = '#ff79c6'
        wrapper.style.color = '#282a36'
        wrapper.style.borderRadius = '4px'
        wrapper.style.fontFamily = 'sans-serif'
        wrapper.style.fontWeight = 'bold'
        wrapper.style.pointerEvents = 'auto' // Re-enable clicks here
        wrapper.style.boxShadow = '0 2px 5px rgba(0,0,0,0.5)'

        // Use React 18 createRoot
        const root = ReactDOM.createRoot(wrapper)

        // Define a simple functional React component
        const App = () => {
            const [count, setCount] = React.useState(0)
            return React.createElement(
                'div',
                { onClick: () => setCount(count + 1), style: { cursor: 'pointer' } },
                `Hyper-Style React Overlay (Clicks: ${count})`
            )
        }

        root.render(React.createElement(App))

        return wrapper
    })
}
