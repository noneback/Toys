import jieba
import numpy
import re
import os
import json
from collections import defaultdict

spam_file_num=7775 
normal_file_num=7063

#todo 原始邮件数据的预处理
'''
获取中文的停用词表
【停用词】（https://zh.wikipedia.org/wiki/%E5%81%9C%E7%94%A8%E8%AF%8D）
'''

def get_stopwords():
    return [
        i.strip() for i in open('./data/中文停用词表.txt', encoding='gbk')
    ]


'''
    读取原始文件，进行简单的处理和分词，返回分词列表
    
'''


def get_raw_str_list(path):
    stop_list = get_stopwords()
    with open(path, encoding='gbk') as f:
        raw_str = f.read()
    pattern = '[^\u4E00-\u9FA5]'  #中文的unicode编码的范围
    regex = re.compile(pattern)
    ret = regex.findall(raw_str)
    handled_str = re.sub(pattern, '', raw_str)
    str_list = [
        word for word in jieba.cut(handled_str) if word not in stop_list
    ]
    return str_list

#分词以及统计词频，得到词汇表
'''
返回一个字典，记录了分词后的词汇表
path:k可以使dir或者file路径，默认为存放文本的dir
is_file_path:表示输入路径是不是文件

'''


def get_voca(path, is_file_path=False):
    if is_file_path:
        return read_voca_from_file(path)

    voca = defaultdict(int)
    # 获取垃圾邮件列表
    file_list = [file for file in os.listdir(path)]
    #获得词汇表
    for file in file_list:  #测试用
        raw_str_list = get_raw_str_list(path + str(file))
        for raw_str in raw_str_list:
            voca[raw_str] = voca[raw_str] + 1

    return voca



'''
    将得到的词汇表保存到json文件中，以便下次读取
    voca:词汇表字典
    path:文件的保存路径
    sort_by_value:是否按value值排序

'''


def save_voca2json(voca, path, sort_by_value=False,indent_=4):
    if sort_by_value:
        sorted_by_value(voca)
        
    with open(path, 'w+') as f:
        f.write(json.dumps(voca, ensure_ascii=False, indent=indent_))


'''
从json文件中读取voca
'''

def read_voca_from_file(path):
    voca = None
    with open(path) as f:
        voca = json.load(f)
    return voca


'''
将字典基于value排序
'''

def sorted_by_value(_dict):
    _dict = dict(sorted(spam_voca.items(), key=lambda x: x[1], reverse=True))
    

#计算 P(Spam|word)

'''
    计算邮件和邮件分类最相关词语及其 P(spam|word)
    words_size:最终使用词语的数量，用于最后的预测
    
'''

def get_top_words_prob(path,spam_voca,normal_voca,words_size=30):
    critical_words=[]
    for word in get_raw_str_list(path):
        if word in spam_voca.keys() and word in normal_voca.keys():
            # 如果word在两边都出现
            p_w_s=spam_voca[word]/spam_file_num
            p_w_n=normal_voca[word]/normal_file_num
            p_s_w=p_w_s/(p_w_n+p_w_s)
            
        elif word in spam_voca.keys() and word not in normal_voca.keys():
            # 如果word只在spam出现
            p_w_s=spam_voca[word]/spam_file_num
            p_w_n=0.01
            p_s_w=p_w_s/(p_w_n+p_w_s)
            
        elif word not in spam_voca.keys() and word in normal_voca.keys():
            # 如果word只在normal出现
            p_w_s=0.01
            p_w_n=normal_voca[word]/normal_file_num
            p_s_w=p_w_s/(p_w_n+p_w_s)
        else:
            #如果都不出现
            p_s_w=0.4
        #以上设定的数值均是基于前人的研究得到的数据
        critical_words.append([word,p_s_w])
        
    return dict(sorted(critical_words[:words_size],key=lambda x:x[1],reverse=True))

'''
    计算贝叶斯概率
    words_prob：包含词汇以及其P(spam|word)概率的一个字典
    spam_voca:垃圾邮件词汇表
    normal_voca:正常邮件词汇表
'''

        
def caculate_bayes(words_prob,spam_voca,normal_voca):
    p_s_w=1
    p_s_nw=1
    for word,prob in words_prob.items():
        p_s_w*=prob
        p_s_nw*=(1-prob)
        
    return p_s_w/(p_s_w+p_s_nw)

def predict(bayes,threshold=0.9):
    return bayes>=threshold    


'''
    返回文件名和label
'''
    
def get_files_labels(dir_path,is_spam=True):
    raw_files_list=os.listdir(dir_path)
    files_list= [dir_path+file for file in raw_files_list]
    labels=[is_spam for i in range(len(files_list))]
    return files_list,labels


'''
    预测测试集结果并打印
    file_list：测试集文件路径集合
    y:测试集结果标签
    word_size：预测的指标
'''

def predict_result(file_list,y,spam_voca,normal_voca,word_size=30):
    ret=[]
    right=0
    for file in file_list:
#         raw_strs=get_raw_str_list(str(file))
        words_prob=get_top_words_prob(file,spam_voca,normal_voca,words_size=word_size)
        bayes=caculate_bayes(words_prob,spam_voca,normal_voca)
        ret.append(predict(bayes))
    for i in range(len(ret)):
        if ret[i]==y[i]:
            right+=1
           
    print(right/len(y))
    

        