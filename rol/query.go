// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rol

const q1 = `( record.city EQUALS "Rockville" ) AND ( record.state EQUALS "MD" )`
const q2 = `( ( record.city EQUALS "Rockville" ) OR ( record.city EQUALS "Gaithersburg" ) ) AND ( record.state EQUALS "MD" )`
