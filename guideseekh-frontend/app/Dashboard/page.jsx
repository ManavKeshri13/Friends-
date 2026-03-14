"use client";
import React, { useState, useEffect } from "react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { Sparkles, BookOpen, ExternalLink, Video, FileText, CheckCircle2, Clock } from "lucide-react";
import ActivityGrid from "../components/ActivityGrid";
import AppSidebar from "../components/AppSidebar";

const A = "oklch(64.6% 0.222 41.116)";

export default function Dashboard() {
  const router = useRouter();
  const [selected, setSelected] = useState(null);
  const [userId, setUserId] = useState(null);
  const [dashboardData, setDashboardData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [quizSubmitted, setQuizSubmitted] = useState(false);
  const [userName, setUserName] = useState("");

  const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || "http://127.0.0.1:8080";

  useEffect(() => {
    const storedUserId = localStorage.getItem("userId");
    const storedUserName = localStorage.getItem("username");
    if (!storedUserId) {
      router.push("/Login");
      return;
    }
    setUserId(storedUserId);
    if (storedUserName) setUserName(storedUserName);
    loadDashboardData(storedUserId);
  }, [router]);

  const loadDashboardData = async (uid) => {
    try {
      const res = await fetch(`${apiBaseUrl}/api/dashboard/${uid}`);
      if (res.ok) {
        const data = await res.json();
        setDashboardData(data);
      }
    } catch (err) {
      console.error("Failed to load dashboard data:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleQuizSubmit = () => {
    if (selected === null) return;
    setQuizSubmitted(true);
    // Optional: Send answer to backend to record daily quiz completion
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-black text-white">
        <div className="flex gap-1.5">
          {[0, 0.15, 0.3].map((d, i) => (
            <span key={i} className={`w-2.5 h-2.5 bg-[#FF5500] rounded-full animate-bounce`} style={{ animationDelay: `${d}s` }} />
          ))}
        </div>
      </div>
    );
  }

  const timetable = dashboardData?.timetable || [];
  const resources = dashboardData?.resources || [];
  const dailyQuiz = dashboardData?.daily_quiz;
  const progress = dashboardData?.progress || { courses_completed: { completed: 0, total: 0 }, quizzes_passed: { passed: 0, total: 0 } };
  const activityData = dashboardData?.activity || [];

  return (
    <div className="min-h-screen text-white flex bg-black">
      <AppSidebar />

      {/* Main */}
      <div className="flex-1 flex flex-col min-w-0 relative overflow-hidden">
        
        {/* Decorative Background Elements */}
        <div className="absolute top-[-20%] left-[-10%] w-[50%] h-[50%] rounded-full opacity-[0.03] blur-[120px] pointer-events-none" style={{ background: A }}></div>
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] rounded-full opacity-[0.02] blur-[100px] pointer-events-none" style={{ background: "white" }}></div>

        {/* Mobile header */}
        <header className="md:hidden flex items-center justify-between px-6 py-4 sticky top-0 z-10 bg-black/60 backdrop-blur-xl border-b border-white/[0.05]">
          <span className="text-lg font-bold text-white tracking-tight">Dashboard</span>
        </header>

        <main className="flex-1 p-4 sm:p-6 lg:p-8 xl:p-12 z-0 overflow-y-auto w-full max-w-[1600px] mx-auto">
          <div className="mb-10">
            <h1 className="text-4xl sm:text-5xl font-extrabold mb-2 tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-white to-gray-400">
              {userName ? `Welcome back, ${userName}` : "Dashboard"}
            </h1>
            <p className="text-gray-400 font-medium tracking-wide">Here&apos;s an overview of your learning journey today.</p>
          </div>

          <div className="grid grid-cols-1 xl:grid-cols-2 gap-6 lg:gap-8">
            {/* Timetable */}
            <div className="rounded-2xl p-6 relative group overflow-hidden transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.12)]" 
                 style={{ background: "linear-gradient(145deg, rgba(255,255,255,0.03) 0%, rgba(255,255,255,0.01) 100%)", border: "1px solid rgba(255,255,255,0.06)" }}>
              <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" 
                   style={{ background: `radial-gradient(800px circle at top center, ${A}15, transparent 40%)` }} />
              
              <div className="flex items-center justify-between mb-6 relative z-10">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-white/5 border border-white/10">
                    <Clock className="w-5 h-5 text-gray-300" />
                  </div>
                  <h3 className="text-xl font-bold text-white tracking-tight">Upcoming Timetable</h3>
                </div>
                <span className="text-xs px-3 py-1.5 rounded-full font-semibold" 
                      style={{ background: `color-mix(in srgb, ${A} 15%, transparent)`, color: A }}>
                  {timetable.length} Scheduled
                </span>
              </div>
              
              {timetable.length === 0 ? (
                <div className="py-12 text-center flex flex-col items-center justify-center rounded-xl bg-black/20 border border-white/5 relative z-10">
                  <div className="w-16 h-16 rounded-full flex items-center justify-center mb-4 bg-white/5 border border-white/10">
                    <Clock className="w-8 h-8 opacity-40 text-gray-400" />
                  </div>
                  <p className="text-gray-300 font-medium">No upcoming schedules.</p>
                  <p className="text-sm mt-1.5 text-gray-500">Set reminders in your chats to stay on track.</p>
                </div>
              ) : (
                <div className="space-y-4 relative z-10">
                  {timetable.map((item, i) => {
                    const dateObj = new Date(item.scheduled_time);
                    const dateStr = dateObj.toLocaleDateString("en-US", { weekday: 'short', month: 'short', day: 'numeric' });
                    const timeStr = dateObj.toLocaleTimeString("en-US", { hour: 'numeric', minute: '2-digit' });
                    
                    return (
                      <div key={item.id} className="flex gap-4 items-center p-4 rounded-xl bg-black/40 border border-white/5 hover:border-white/15 transition-all group/item cursor-pointer"
                           onClick={() => router.push("/chat")}>
                        <div className="w-1.5 h-12 rounded-full" style={{ background: A }}></div>
                        <div className="flex-1 min-w-0">
                          <p className="text-base font-semibold text-white truncate group-hover/item:text-transparent group-hover/item:bg-clip-text group-hover/item:bg-gradient-to-r group-hover/item:from-white group-hover/item:to-gray-300 transition-all">{item.topic || "Study Session"}</p>
                          <div className="flex items-center gap-2 mt-1">
                            <span className="text-xs font-medium text-gray-400 tracking-wide uppercase">{dateStr} • {timeStr}</span>
                          </div>
                        </div>
                        <div className="w-8 h-8 rounded-full bg-white/5 flex items-center justify-center group-hover/item:bg-white/20 transition-all">
                          <ExternalLink className="w-4 h-4 text-gray-400 group-hover/item:text-white" />
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>

            {/* Recent Resources */}
            <div className="rounded-2xl p-6 relative group overflow-hidden transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.12)]" 
                 style={{ background: "linear-gradient(145deg, rgba(255,255,255,0.03) 0%, rgba(255,255,255,0.01) 100%)", border: "1px solid rgba(255,255,255,0.06)" }}>
              <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" 
                   style={{ background: `radial-gradient(800px circle at top center, ${A}15, transparent 40%)` }} />
                   
              <div className="flex items-center justify-between mb-6 relative z-10">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-white/5 border border-white/10">
                    <BookOpen className="w-5 h-5 text-gray-300" />
                  </div>
                  <h3 className="text-xl font-bold text-white tracking-tight">Recent Resources</h3>
                </div>
                <button onClick={() => router.push("/chat")} className="text-xs font-medium text-gray-400 hover:text-white transition-colors">View All</button>
              </div>
              
              {resources.length === 0 ? (
                <div className="py-12 text-center flex flex-col items-center justify-center rounded-xl bg-black/20 border border-white/5 relative z-10">
                  <div className="w-16 h-16 rounded-full flex items-center justify-center mb-4 bg-white/5 border border-white/10">
                    <BookOpen className="w-8 h-8 opacity-40 text-gray-400" />
                  </div>
                  <p className="text-gray-300 font-medium">No resources found.</p>
                  <p className="text-sm mt-1.5 text-gray-500">Chat with the AI to generate tailored study materials.</p>
                </div>
              ) : (
                <div className="space-y-3 relative z-10">
                  {resources.map((res, i) => {
                    const type = (res.resource_type || res.type || "").toLowerCase();
                    const IconObj = type.includes("video") ? Video :
                                    type.includes("doc") ? FileText :
                                    ExternalLink;
                    const title = res.resource_title || res.title || "Resource";
                    const url = res.resource_url || res.url || "#";
                    
                    return (
                      <a key={res.id || i} href={url} target="_blank" rel="noopener noreferrer" 
                        className="flex items-center gap-4 p-4 rounded-xl bg-white/[0.02] border border-white/[0.04] hover:bg-white/[0.06] hover:border-white/[0.1] hover:scale-[1.01] transition-all duration-300 group/res">
                        <div className="p-2.5 rounded-lg bg-[#111] border border-white/10 text-gray-400 group-hover/res:text-white group-hover/res:border-white/20 transition-all shadow-inner">
                          <IconObj className="w-5 h-5" />
                        </div>
                        <div className="min-w-0 flex-1">
                          <p className="text-sm sm:text-base font-semibold text-gray-200 group-hover/res:text-white transition-colors line-clamp-1">{title}</p>
                          <span className="text-[10px] uppercase tracking-widest text-gray-500 font-bold mt-1 block truncate group-hover/res:text-gray-400 transition-colors" style={{ color: A }}>{type || "resource"}</span>
                        </div>
                        <ExternalLink className="w-4 h-4 opacity-0 -translate-x-2 group-hover/res:opacity-100 group-hover/res:translate-x-0 transition-all duration-300 text-gray-400" />
                      </a>
                    )
                  })}
                </div>
              )}
            </div>

            {/* Quiz */}
            <div className="rounded-2xl p-6 flex flex-col relative group overflow-hidden transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.12)]" 
                 style={{ background: "linear-gradient(145deg, rgba(255,255,255,0.03) 0%, rgba(255,255,255,0.01) 100%)", border: "1px solid rgba(255,255,255,0.06)" }}>
              <div className="absolute top-0 right-0 w-32 h-32 bg-white opacity-[0.02] blur-3xl pointer-events-none rounded-full blur-[80px]" style={{ background: A }}></div>
              
              <div className="flex items-center justify-between mb-6 relative z-10">
                <div className="flex items-center gap-3">
                  <div className="p-2 rounded-lg bg-gradient-to-br from-white/10 to-transparent border border-white/10">
                    <Sparkles className="w-5 h-5" style={{ color: A }} />
                  </div>
                  <h3 className="text-xl font-bold text-white tracking-tight">Daily Challenge</h3>
                </div>
              </div>
              
              {!dailyQuiz ? (
                <div className="flex-1 flex flex-col items-center justify-center py-10 text-center rounded-xl bg-black/20 border border-white/5 relative z-10">
                  <div className="w-16 h-16 rounded-full flex items-center justify-center mb-4 bg-green-500/10 border border-green-500/20 shadow-[0_0_30px_rgba(34,197,94,0.1)]">
                    <CheckCircle2 className="w-8 h-8 text-green-500" />
                  </div>
                  <p className="text-gray-200 font-bold text-lg mb-1">You're all caught up!</p>
                  <p className="text-sm text-gray-500">Your daily learning objective is complete.</p>
                </div>
              ) : quizSubmitted ? (
                 <div className="flex-1 flex flex-col items-center justify-center py-10 text-center text-white rounded-xl bg-black/20 border border-white/5 relative z-10">
                  <div className="w-16 h-16 rounded-full flex items-center justify-center mb-4 bg-green-500/10 border border-green-500/20 shadow-[0_0_30px_rgba(34,197,94,0.1)] scale-in-center">
                    <CheckCircle2 className="w-8 h-8 text-green-500" />
                  </div>
                  <h4 className="font-bold text-xl mb-2 tracking-tight">Answer Recorded!</h4>
                  <p className="text-sm text-gray-400 font-medium">Consistency is key to mastery.</p>
                </div>
              ) : (
                <div className="flex flex-col flex-1 relative z-10">
                  <div className="p-5 rounded-xl bg-white/[0.03] border border-white/10 mb-5 shadow-inner">
                    <p className="text-[15px] text-gray-100 font-medium leading-relaxed text-left">
                      {dailyQuiz.question}
                    </p>
                  </div>
                  
                  <div className="space-y-3 mb-6 flex-1 text-left">
                    {(() => {
                      let options = [];
                      try { options = JSON.parse(dailyQuiz.options || "[]"); } catch (e) { options = [dailyQuiz.answer]; }
                      return options.map((text, idx) => (
                        <label key={idx} onClick={() => setSelected(idx)}
                          className="flex items-center gap-4 p-4 rounded-xl border cursor-pointer transition-all duration-200 relative overflow-hidden group/opt"
                          style={{ 
                            borderColor: selected === idx ? A : "rgba(255,255,255,0.08)", 
                            background: selected === idx ? `color-mix(in srgb, ${A} 10%, transparent)` : "rgba(255,255,255,0.02)" 
                          }}>
                          
                          <div className={`w-5 h-5 rounded-full border flex items-center justify-center transition-all ${selected === idx ? "border-transparent" : "border-gray-500 group-hover/opt:border-gray-300"}`}
                               style={{ background: selected === idx ? A : "" }}>
                            {selected === idx && <div className="w-2 h-2 rounded-full bg-black"></div>}
                          </div>
                          
                          <span className={`text-[15px] font-medium transition-colors ${selected === idx ? "text-white" : "text-gray-300 group-hover/opt:text-white"}`}>
                            {text}
                          </span>
                        </label>
                      ));
                    })()}
                  </div>
                  
                  <button onClick={handleQuizSubmit} disabled={selected === null}
                    className="w-full py-4 rounded-xl font-bold text-sm text-[oklch(97%_0_0)] shadow-lg transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed hover:-translate-y-0.5 mt-auto relative overflow-hidden group/btn"
                    style={{ background: A }}>
                    <div className="absolute inset-0 bg-white/20 translate-y-full group-hover/btn:translate-y-0 transition-transform duration-300 ease-out"></div>
                    <span className="relative z-10">Submit Answer</span>
                  </button>
                </div>
              )}
            </div>

            {/* Progress */}
            <div className="rounded-2xl p-6 relative group overflow-hidden transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.12)] flex flex-col" 
                 style={{ background: "linear-gradient(145deg, rgba(255,255,255,0.03) 0%, rgba(255,255,255,0.01) 100%)", border: "1px solid rgba(255,255,255,0.06)" }}>
              <div className="absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none" 
                   style={{ background: `radial-gradient(800px circle at bottom center, ${A}15, transparent 40%)` }} />
                   
              <h3 className="text-xl font-bold mb-6 text-white tracking-tight relative z-10">My Progress</h3>
              
              <div className="space-y-6 mb-8 relative z-10">
                {[
                  { 
                    label: "Quizzes Attempted", 
                    value: `${progress.courses_completed.completed}`, 
                    pct: progress.courses_completed.total > 0 ? Math.round((progress.courses_completed.completed / progress.courses_completed.total) * 100) : 0 
                  },
                  { 
                    label: "Quizzes Passed (80%+)", 
                    value: `${progress.quizzes_passed.passed}`, 
                    pct: progress.quizzes_passed.total > 0 ? Math.round((progress.quizzes_passed.passed / progress.quizzes_passed.total) * 100) : 0 
                  }
                ].map(({ label, value, pct }) => (
                  <div key={label} className="group/prog">
                    <div className="flex justify-between mb-2.5">
                      <span className="text-sm font-medium text-gray-300 group-hover/prog:text-white transition-colors">{label}</span>
                      <span className="text-sm font-bold text-white tracking-wide">
                        {value} <span className="text-gray-500 font-medium ml-1 bg-white/10 px-1.5 py-0.5 rounded-md text-xs">{pct}%</span>
                      </span>
                    </div>
                    <div className="w-full h-3 rounded-full bg-black/40 border border-white/5 overflow-hidden shadow-inner">
                      <div className="h-full rounded-full transition-all duration-1000 ease-out relative" style={{ width: `${pct}%`, background: A }}>
                        <div className="absolute top-0 right-0 bottom-0 left-0 bg-gradient-to-r from-transparent to-white/30" />
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              
              <div className="mt-auto relative z-10 pt-4 border-t border-white/5">
                <ActivityGrid data={activityData} />
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
