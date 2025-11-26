import React from "react";
import './Pane.css';

interface PaneProps {
    children: JSX.Element[] | JSX.Element;
    layout: 'single' | 'vsplit' | 'hsplit';
}

export function Pane({
    children,
    layout,
}: PaneProps) {
    if (layout === 'single') {
        if (React.Children.count(children) === 0) {
            return <div className='pane-single'>"Nothing to see here"</div>;
        }
        return <div className='pane-single'>{children}</div>;
    }

    return (
        <div className={`pane-${layout}`}>
            {React.Children.map(children, c => (
                <div className={`pane-${layout}-child`}>{c}</div>
            ))}
        </div>
    );
}

