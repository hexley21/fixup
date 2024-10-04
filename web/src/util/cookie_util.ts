// export function getCookie(name: string) {
//   var dc = document.cookie;
//   var prefix = name + "=";
//   var begin = dc.indexOf("; " + prefix);
  
//   if (begin !== -1) {
//     begin += 2;
//     var end = document.cookie.indexOf(";", begin);
//     if (end === -1) {
//       end = dc.length;
//     }
//   } else {
//     begin = dc.indexOf(prefix);
//     if (begin !== 0) return null;
//   }

//   return decodeURI(dc.substring(begin + prefix.length, end));
// }