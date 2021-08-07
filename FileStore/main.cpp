#include <bits/stdc++.h>

using namespace std;

class Solution {
private:
  unordered_set<vector<int>> s;
  vector<int> mid;
  int val = 0;
  int sum = 0;

public:
  Solution() {}
  void helper(int n, int k) {
    if (mid.size() == k) {
      if (sum == n) {
        sort(mid.begin(), mid.end());
        s.insert(mid);
        return;
      }
    }
    if (sum > n)
      return;

    val++;
    int tmp = val;
    mid.push_back(val);
    sum += val;
    val = 0;
    helper(n, k);
    mid.pop_back();
    val = tmp;
    sum -= val;

    helper(n, k);
  }

  int divide(int n, int k) {
    // write code here
    if (n < k)
      return 0;
    if (n == k)
      return 1;

    helper(n, k);
    return s.size();
  }
};

int main() {
  Solution st{};

  cout << st.divide(7, 3);
}
