export function ContentLayout({ children }: { children: React.ReactNode }) {
    return (
        <div className="max-w-7xl m-auto backgr bg-white">
            {children}
        </div>
    );
}
