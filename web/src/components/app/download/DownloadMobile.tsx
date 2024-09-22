export default function DownloadMobile() {
  return (
    <div className="w-full max-w-xl">
      <h3 className="text-3xl text-left">fixup.com — thousands of specialists in your smartphone</h3>
      <p className="text-xl text-muted-foreground mt-2 text-left">Download the free app and order any service on the go</p>
      <div className="flex items-center text-sm justify-center mt-8 gap-4">
        <a href="#" className="hover:underline">
          <img src="../../../assets/badges/google_play.svg" className="h-14" />
        </a>
        <a href="#" className="hover:underline">
          <img src="../../../assets/badges/app_store.svg" className="h-14" />
        </a>
      </div>
    </div>
  )
}