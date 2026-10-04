export function DefaultBrand() {
  return (
    <span className="block h-11 w-[134px]" role="img" aria-label="GooseForum">
      <img src="/static/brand/gooseforum-light.svg" alt="" width={1898} height={625} className="h-full w-full object-contain dark:hidden" />
      <img src="/static/brand/gooseforum-dark.svg" alt="" width={1898} height={625} className="hidden h-full w-full object-contain dark:block" />
    </span>
  )
}
