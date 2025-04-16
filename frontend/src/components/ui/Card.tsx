import React from 'react';
import styles from './Card.module.css';
import classNames from 'classnames';

interface CardProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function Card({ children, className, id }: CardProps) {
    return (
        <div id={id} className={classNames(styles.card, className)}>
            {children}
        </div>
    );
}

interface CardHeaderProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function CardHeader({ children, className, id }: CardHeaderProps) {
    return (
        <div id={id} className={classNames(styles.cardHeader, className)}>
            {children}
        </div>
    );
}

interface CardTitleProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function CardTitle({ children, className, id }: CardTitleProps) {
    return (
        <h3 id={id} className={classNames(styles.cardTitle, className)}>
            {children}
        </h3>
    );
}

interface CardDescriptionProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function CardDescription({ children, className, id }: CardDescriptionProps) {
    return (
        <p id={id} className={classNames(styles.cardDescription, className)}>
            {children}
        </p>
    );
}

interface CardContentProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function CardContent({ children, className, id }: CardContentProps) {
    return (
        <div id={id} className={classNames(styles.cardContent, className)}>
            {children}
        </div>
    );
}

interface CardFooterProps {
    children: React.ReactNode;
    className?: string;
    id?: string;
}

export function CardFooter({ children, className, id }: CardFooterProps) {
    return (
        <div id={id} className={classNames(styles.cardFooter, className)}>
            {children}
        </div>
    );
}
