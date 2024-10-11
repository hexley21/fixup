export function trimStart(character: string, value: string) {
    let startIndex = 0;
  
    while (value[startIndex] === character) {
      startIndex++;
    }
  
    return value.substr(startIndex);
  }
  
  export function trimEnd(character: string, value: string) {
    return reverse(trimStart(character, reverse(value)));
  }
  
  function reverse(value: string) {
    return value
      .split("")
      .reverse()
      .join("");
  }